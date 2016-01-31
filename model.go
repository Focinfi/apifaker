package apifaker

import (
	"encoding/json"
	"fmt"
	"github.com/Focinfi/gset"
	"github.com/Focinfi/gtester"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/inflection"
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

	// relationship
	HasMany []string `json:"has_many"`
	HasOne  []string `json:"has_one"`

	// Set Compose *gset.SetThreadSafe to contains data
	Set *gset.SetThreadSafe `json:"-"`

	// CurrentId record the number of times of adding
	CurrentId float64 `json:"current_id"`

	// dataChanged record the sign if this Model's Set have been changed
	dataChanged bool `json:"-"`
	sync.Mutex  `json:"-"`
	router      *Router `json:"-"`
}

// NewModel allocates and returns a new Model
func NewModel(router *Router) *Model {
	return &Model{
		Seeds:   []map[string]interface{}{},
		Columns: []*Column{},
		Set:     gset.NewSetThreadSafe(),
		router:  router,
	}
}

// nextId plus 1 to Model.CurrentId return return it
func (model *Model) nextId() float64 {
	model.CurrentId++
	return model.CurrentId
}

// Has return if m has LineItem with id param
func (model *Model) Has(id float64) bool {
	return model.Set.Has(gset.T(id))
}

// Get get and return element with id param, also return
// if it's not nil or a LineItem
func (model Model) Get(id float64) (li LineItem, ok bool) {
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

	// set id
	if _, ok := li.Get("id"); !ok {
		li.Set("id", model.nextId())
	}

	if err := model.checkSeed(li.ToMap()); err != nil {
		return err
	} else {
		model.Set.Add(li)
		model.dataChanged = true
		model.addUniqueValues(li)
	}

	return nil
}

// Update a LineItem in Model
func (model *Model) Update(id float64, li *LineItem) error {
	model.Lock()
	defer model.Unlock()

	// set id
	if _, ok := li.Get("id"); !ok {
		li.Set("id", id)
	}

	if err := model.checkSeed(li.dataMap); err != nil {
		return err
	} else {
		if model.Set.Has(gset.T(id)) {
			model.Set.Add(*li)
			model.dataChanged = true
			model.removeUniqueValues(*li)
			model.addUniqueValues(*li)
		} else {
			return fmt.Errorf("model[id:%d] does not exsit", id)
		}
	}

	return nil
}

// Delete
func (model *Model) Delete(id float64) {
	model.Lock()
	defer model.Unlock()

	li, ok := model.Get(id)

	if !ok {
		return
	}

	model.DeleteRelativeLis(id)
	model.Set.Remove(gset.T(id))
	model.dataChanged = true
	model.removeUniqueValues(li)
}

func (model *Model) InsertRelativeData(li *LineItem) error {
	// has one relationship
	singularName := inflection.Singular(model.Name)
	for _, resName := range model.HasOne {
		resStruct := map[string]interface{}{}
		if resRouter, ok := model.router.apiFaker.Routers[resName]; !ok {
			return fmt.Errorf("has no resource %s", resName)
		} else {
			resLis := resRouter.Model.ToLineItems()
			sort.Sort(resLis)
			for _, resLi := range resLis {
				curId, ok := resLi.Get(fmt.Sprintf("%s_id", singularName))
				if ok && curId == li.Id() {
					resStruct = resLi.ToMap()
					break
				}
			}
		}
		if len(resStruct) > 0 {
			li.Set(inflection.Singular(resName), resStruct)
		}
	}

	// has many
	for _, resName := range model.HasMany {
		resSlice := []interface{}{}
		if resRouter, ok := model.router.apiFaker.Routers[resName]; !ok {
			return fmt.Errorf("has no resource %s", resName)
		} else {
			resLis := resRouter.Model.ToLineItems()
			sort.Sort(resLis)
			for _, resLi := range resLis {
				curId, ok := resLi.Get(fmt.Sprintf("%s_id", singularName))
				if ok && curId == li.Id() {
					resSlice = append(resSlice, resLi.ToMap())
				}
			}
		}
		if len(resSlice) > 0 {
			li.Set(resName, resSlice)
		}
	}

	return nil
}

func (model *Model) DeleteRelativeLis(id float64) {
	_, ok := model.Get(id)
	if !ok {
		return
	}

	for _, rotuer := range model.router.apiFaker.Routers {
		var isRelativeRouter bool
		var foreign_key = fmt.Sprintf("%s_id", inflection.Singular(model.Name))
		for _, column := range rotuer.Model.Columns {
			if column.Name == foreign_key {
				isRelativeRouter = true
				break
			}
		}

		if isRelativeRouter {
			for _, li := range rotuer.Model.ToLineItems() {
				key, ok := li.Get(foreign_key)
				fmt.Printf("[Foriegn]%s, %s \n", key, id)
				if ok && key == id {
					rotuer.Model.Delete(li.Id().(float64))
				}
			}
		}
	}
}

// ToLineItems allocate a new LineItems filled with
// Model elements slice
func (model Model) ToLineItems() LineItems {
	lis := []LineItem{}
	elements := model.Set.ToSlice()
	for _, element := range elements {
		if li, ok := element.(LineItem); ok {
			model.InsertRelativeData(&li)
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
func (model *Model) removeUniqueValues(lis ...LineItem) {
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
func (model *Model) UpdateWithAttrsInGinContext(id float64, ctx *gin.Context) (int, interface{}) {
	// check if element does exsit
	li, ok := model.Get(id)
	if !ok {
		return http.StatusNotFound, nil
	}

	// update model
	for _, column := range model.Columns {
		if value := ctx.PostForm(column.Name); value != "" {
			if err := column.CheckValue(value, model); err != nil {
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
func (model *Model) UpdateWithAllAttrsInGinContex(id float64, ctx *gin.Context) (int, interface{}) {
	// check if element does exsit
	_, ok := model.Get(id)
	if !ok {
		return http.StatusNotFound, nil
	}

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

// checkRelationshipsMeta
func (model *Model) checkRelationshipsMeta() error {
	set := gset.NewSetSimple()
	for _, resName := range model.HasMany {
		if set.Has(resName) {
			return fmt.Errorf("%s has been used", resName)
		} else {
			set.Add(resName)
		}
	}
	for _, resName := range model.HasOne {
		if set.Has(resName) {
			return fmt.Errorf("%s has been used", resName)
		} else {
			set.Add(resName)
		}
	}
	return nil
}

// checkColumnsMeta
func (model *Model) checkColumnsMeta() error {
	if len(model.Columns) < 1 || model.Columns[0].Name != "id" {
		return fmt.Errorf("The first colmun must be id")
	}

	for _, column := range model.Columns {
		if err := column.CheckMeta(); err != nil {
			return err
		}
	}

	return nil
}

func (model *Model) checkRelationships(seed map[string]interface{}) error {
	for _, column := range model.Columns {
		if err := column.CheckRelationships(seed[column.Name], model); err != nil {
			return err
		}
	}
	return nil
}

// checkSeed check specific seed
func (model *Model) checkSeedBasic(seed map[string]interface{}) error {
	columns := model.Columns

	if len(seed) != len(columns) {
		return ColumnCountError
	}

	for _, column := range columns {
		if seedVal, ok := seed[column.Name]; !ok {
			ColumnNameError = fmt.Errorf("has not column: %s", column.Name)
			return ColumnNameError
		} else {
			// check seedVal
			if err := column.CheckValue(seedVal, model); err != nil {
				return err
			}
		}
	}

	return nil
}

func (model *Model) checkSeed(seed map[string]interface{}) error {
	return gtester.NewCheckQueue().Add(func() error {
		return model.checkSeedBasic(seed)
	}).Add(func() error {
		return model.checkRelationships(seed)
	}).Run()
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

func (model *Model) checkSeedsUniqueness() error {
	for _, seed := range model.Seeds {
	}

	return nil
}

func (model *Model) checkSeedsRelationships() error {
	for _, seed := range model.Seeds {
		if err := model.checkRelationships(seed); err != nil {
			return err
		}
	}

	return nil
}

func (model *Model) checkSeedsBasic() error {
	for _, seed := range model.Seeds {
		if err := model.checkSeedBasic(seed); err != nil {
			return err
		}
	}

	return nil
}

func (model *Model) initSet() error {
	if err := model.checkSeeds(); err != nil {
		return err
	}
	return nil
}

// setItems
func (model *Model) setItems() {
	if model.Set == nil {
		model.Set = gset.NewSetThreadSafe()
	}
	for _, seed := range model.Seeds {
		li := NewLineItemWithMap(seed)
		model.Set.Add(li)
		model.addUniqueValues(li)
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
func NewModelWithPath(path string, router *Router) (*Model, error) {
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
	model := NewModel(router)
	if err = json.Unmarshal(bytes, model); err != nil {
		err = fmt.Errorf("[apifaker] json format error: %s, file: %s", err.Error(), path)
	} else {
		// use CheckQueue to check columns and seeds
		err = gtester.NewCheckQueue().Add(func() error {
			return model.checkRelationshipsMeta()
			return nil
		}).Add(func() error {
			return model.checkColumnsMeta()
		}).Add(func() error {
			return model.checkSeedsBasic()
		}).Run()

		// set items from seeds for later relationships validation
		if err == nil {
			model.setItems()
		}
	}

	return model, err
}
