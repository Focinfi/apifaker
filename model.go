package apifaker

import (
	"encoding/json"
	"fmt"
	"github.com/Focinfi/gset"
	"github.com/Focinfi/gtester"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
	"os"
	"sort"
	"sync"
)

type Model struct {
	Name string `json:"resource_name"`

	// Seeds contains value of "seeds" in json files
	Seeds   []map[string]interface{} `json:"seeds"`
	Columns []*Column                `json:"columns"`

	// Set Compose *gset.SetThreadSafe to contains data
	Set *gset.SetThreadSafe `json:"-"`

	// currentId record the number of times of adding
	currentId int `json:"-"`

	// dataChanged record the sign if this Model's Set have been changed
	dataChanged bool `json:"-"`
	sync.Mutex  `json:"-"`
}

// NewModel allocates and returns a new Model
func NewModel() *Model {
	return &Model{
		Seeds:   []map[string]interface{}{},
		Columns: []*Column{},
		Set:     gset.NewSetThreadSafe(),
	}
}

// nextId plus 1 to Model.currentId return return it
func (model *Model) nextId() int {
	model.currentId++
	return model.currentId
}

// Has return if m has LineItem with id param
func (model *Model) Has(id int) bool {
	return model.Set.Has(gset.T(id))
}

// Get get and return element with id param, also return
// if it's not nil or a LineItem
func (model Model) Get(id int) (li LineItem, ok bool) {
	var element interface{}
	if element, ok = model.Set.Get(id); ok {
		li, ok = element.(LineItem)
	}
	return
}

// Add add a LineItem to Model.Set
func (model *Model) Add(li LineItem) error {
	model.Lock()
	defer model.Unlock()

	if err := model.checkSeed(li.ToMap()); err != nil {
		return err
	} else {
		li.Set("id", model.nextId())
		model.Set.Add(li)
		model.dataChanged = true
		model.addUniqueValues(li)
	}

	return nil
}

// Update a LineItem in Model
func (model *Model) Update(id int, li *LineItem) error {
	model.Lock()
	defer model.Unlock()

	if err := model.checkSeed(li.dataMap); err != nil {
		return err
	} else {
		if model.Set.Has(gset.T(id)) {
			li.Set("id", id)
			model.Set.Add(li)
			model.dataChanged = true
			model.removeUniqueValues(li)
			model.addUniqueValues(*li)
		} else {
			return fmt.Errorf("model[id:%d] does not exsit", id)
		}
	}

	return nil
}

// Delete
func (model *Model) Delete(id int) {
	model.Lock()
	defer model.Unlock()

	value, _ := model.Set.Get(id)

	model.Set.Remove(gset.T(id))
	model.dataChanged = true
	model.removeUniqueValues(value.(*LineItem))
}

// ToLineItems allocate a new LineItems filled with
// Model elements slice
func (model Model) ToLineItems() LineItems {
	lis := []LineItem{}
	elements := model.Set.ToSlice()
	for _, element := range elements {
		if li, ok := element.(LineItem); ok {
			lis = append(lis[:], li)
		}
	}
	return LineItems(lis)
}

// addUniqueValues add values of Lineitem into corresponding Column's uniqueValues
func (model *Model) addUniqueValues(lis ...LineItem) {
	for _, li := range lis {
		for _, column := range model.Columns {
			if value, ok := li.Get(column.Name); ok {
				column.AddValue(value)
			}
		}
	}
}

// removeUniqueValues reomve values of Lineitem into corresponding Column's uniqueValues
func (model *Model) removeUniqueValues(lis ...*LineItem) {
	for _, li := range lis {
		for _, column := range model.Columns {
			if value, ok := li.Get(column.Name); ok {
				column.RemoveValue(value)
			}
		}
	}
}

// UpdateWithAttrsInGinContext find a LineItem with id param,
// update it with attrs from gin.Contex.PostForm(),
// returns status in net/http package and object for response
func (model *Model) UpdateWithAttrsInGinContext(id int, ctx *gin.Context) (int, interface{}) {
	// check if element does exsit
	li, ok := model.Get(id)
	if !ok {
		return http.StatusNotFound, nil
	}

	// update model
	for _, column := range model.Columns {
		if value := ctx.PostForm(column.Name); value != "" {
			if err := column.CheckValue(value); err != nil {
				return http.StatusBadRequest, map[string]interface{}{"message": err.Error()}
			} else {
				li.Set(column.Name, value)
			}
		}
	}
	return http.StatusOK, li.ToMap()
}

// UpdateWithAllAttrsInGinContex find a LineItem with id param,
// allocate a new LineItem with attrs from gin.Context.PostForm(),
// replace this LineItem with the new one,
// returns status in net/http package, and object for response
func (model *Model) UpdateWithAllAttrsInGinContex(id int, ctx *gin.Context) (int, interface{}) {
	// check if element does exsit
	_, ok := model.Get(id)
	if !ok {
		return http.StatusNotFound, nil
	}
	fmt.Println("[Update found]", ok)

	// update this li with a new item
	var newItem LineItem
	err := gtester.NewCheckQueue().Add(func() error {
		var err error
		newItem, err = NewLineItemWithGinContext(ctx, model)
		return err
	}).Add(func() error {
		return model.Update(id, &newItem)
	}).Run()

	// check err
	if err != nil {
		return http.StatusBadRequest, map[string]string{"message": err.Error()}
	} else {
		return http.StatusOK, newItem.ToMap()
	}
}

// checkColumnsMeta
func (model *Model) checkColumnsMeta() error {
	for _, column := range model.Columns {
		if err := column.CheckMeta(); err != nil {
			return err
		}
	}

	return nil
}

// checkSeed check specific seed
func (model *Model) checkSeed(seed map[string]interface{}) error {
	columns := model.Columns
	delete(seed, "id")

	if len(seed) != len(columns) {
		return ColumnCountError
	}

	for _, column := range columns {
		if seedVal, ok := seed[column.Name]; !ok {
			ColumnNameError = fmt.Errorf("has not column: %s", column.Name)
			return ColumnNameError
		} else {
			// check seedVal
			return column.CheckValue(seedVal)
		}
	}

	return nil
}

// checkSeeds() check if every item of this Model object's Seeds
// is in line of description of its Columns.
func (model *Model) checkSeeds() error {
	for _, seed := range model.Seeds {
		if err := model.checkSeed(seed); err != nil {
			return err
		}
	}

	return nil
}

// setItems
func (model *Model) setItems() {
	if model.Set == nil {
		model.Set = gset.NewSetThreadSafe()
	}
	for _, seed := range model.Seeds {
		model.Add(NewLineItemWithMap(seed))
	}
}

// SetToSeeds
func (model *Model) SetToSeeds() {
	models := model.ToLineItems()
	sort.Sort(models)
	model.Seeds = models.ToSlice()
}

// SaveToFile save model to file with the given path
func (model *Model) SaveToFile(path string) error {
	if !model.dataChanged {
		return nil
	}

	model.Lock()
	defer model.Unlock()

	// open file
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	// set m.Set to m.seeds
	model.SetToSeeds()
	bytes, err := json.Marshal(model)
	if err != nil {
		return err
	}
	_, err = file.WriteString(string(bytes))

	// set dataChanged to false
	if err == nil {
		model.dataChanged = false
	}

	return err
}

// NewModelWithPath allocates and returns a new Model,
// using the given path as it's json file path
func NewModelWithPath(path string) (*Model, error) {
	// open file
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// read and unmarshal file
	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}
	model := NewModel()
	if err = json.Unmarshal(bytes, model); err != nil {
		err = fmt.Errorf("[apifaker] json format error: %s, file: %s", err.Error(), path)
	} else {
		// use CheckQueue to check columns and seeds
		err = gtester.NewCheckQueue().Add(func() error {
			return model.checkSeeds()
		}).Add(func() error {
			return model.checkColumnsMeta()
		}).Run()

		if err == nil {
			model.setItems()
			model.addUniqueValues(model.ToLineItems()...)
		}
	}

	return model, err
}
