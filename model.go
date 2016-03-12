package apifaker

import (
	"encoding/json"
	"github.com/Focinfi/gset"
	"github.com/Focinfi/gtester"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/inflection"
	"io/ioutil"
	"os"
	"sort"
	"sync"
)

type Model struct {
	Name string `json:"resource_name"`

	// Seeds acts as a snapshot of the whole database
	Seeds   []map[string]interface{} `json:"seeds"`
	Columns []*Column                `json:"columns"`

	// relationships
	HasMany []string `json:"has_many"`
	HasOne  []string `json:"has_one"`

	// Set contains runtime data
	Set *gset.SetThreadSafe `json:"-"`

	// currentId records the max of id
	currentId float64

	// dataChanged signs if differs between Set and Seeds
	dataChanged bool
	sync.RWMutex
	router *Router
}

//------Model CURD------//
// NewModel allocates and returns a new Model
func NewModel(router *Router) *Model {
	return &Model{
		Seeds:   []map[string]interface{}{},
		Columns: []*Column{},
		Set:     gset.NewSetThreadSafe(),
		router:  router,
	}
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

	model := NewModel(router)
	bytes := []byte{}

	err = gtester.NewInspector().Check(func() error {
		bytes, err = ioutil.ReadAll(file)
		return err
	}).Check(func() error {
		return json.Unmarshal(bytes, model)
	}).Check(func() error {
		return model.CheckRelationshipsMeta()
	}).Check(func() error {
		return model.CheckColumnsMeta()
	}).Check(func() error {
		return model.ValidateSeedsValue()
	}).Then(func() {
		model.initSet()
	})

	return model, err
}

// updateId updates currentId if the given id is bigger
func (model *Model) updateId(id float64) {
	if id > model.currentId {
		model.currentId = id
	}
}

// nextId pluses 1 to Model.currentId and returns it
func (model *Model) nextId() float64 {
	model.currentId++
	return model.currentId
}

// Len returns the length of Model's Set
func (model *Model) Len() int {
	return model.Set.Len()
}

// Has returns if Model has LineItem with the given id
func (model *Model) Has(id float64) bool {
	return model.Set.Has(gset.T(id))
}

// Get gets and returns element with id param and the existence of it
func (model *Model) Get(id float64) (li LineItem, ok bool) {
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

	// set id if the given LineItem has no id
	if _, ok := li.Get("id"); !ok {
		li.Set("id", model.nextId())
	}

	if err := model.Validate(li.ToMap()); err != nil {
		return err
	} else {
		model.Set.Add(li)
		model.dataChanged = true
		model.addUniqueValues(li)
	}

	return nil
}

// Update updates the LineItem with the given id by the given LineItem
func (model *Model) Update(id float64, li *LineItem) error {
	oldLi, ok := model.Get(id)
	if !ok {
		return SeedsErrorf("model %s[id:%d] does not exsit", model.Name, id)
	}

	model.Lock()
	defer model.Unlock()

	// set id if the given LineItem has no id
	if _, ok := li.Get("id"); !ok {
		li.Set("id", id)
	}

	if err := model.Validate(li.dataMap); err != nil {
		return err
	} else {
		model.Set.Add(*li)
		model.dataChanged = true
		model.removeUniqueValues(oldLi)
		model.addUniqueValues(*li)
	}

	return nil
}

// Delete deletes the LineItem and its related data with the given id
func (model *Model) Delete(id float64) {
	model.Lock()
	defer model.Unlock()

	li, ok := model.Get(id)

	if !ok {
		return
	}

	li.DeleteRelatedLis(id, model)
	model.Set.Remove(gset.T(id))
	model.dataChanged = true
	model.removeUniqueValues(li)
}

// UpdateWithAttrsInGinContext finds a LineItem with id param,
// updates it with attrs from gin.Contex.PostForm(),
// returns the edited LineItem
func (model *Model) UpdateWithAttrs(id float64, ctx *gin.Context) (LineItem, error) {
	// check if element does exsit
	li, ok := model.Get(id)
	if !ok {
		return li, SeedsErrorf("model %s[id:%d] does not exsit", model.Name, id)
	}

	// update model
	for _, column := range model.Columns {
		value := ctx.PostForm(column.Name)

		if value == "" || column.Name == "id" {
			continue
		}

		formatVal, err := FormatValue(column.Type, value)
		if err == nil {
			err = column.CheckValue(formatVal, model)
		}

		if err != nil {
			return li, err
		} else {
			oldValue, _ := li.Get(column.Name)
			column.RemoveUniquenessOf(oldValue)
			li.Set(column.Name, formatVal)
			column.AddUniquenessOf(formatVal)
		}
	}
	return li, nil
}

//------End Model CURD------//

//------Columns Uniqueness------//
// addUniqueValues adds values of the Lineitem into corresponding Column's uniqueValues
func (model *Model) addUniqueValues(lis ...LineItem) {
	for _, li := range lis {
		for _, column := range model.Columns {
			if value, ok := li.Get(column.Name); ok {
				column.AddUniquenessOf(value)
			}
		}
	}
}

// removeUniqueValues reomves values of Lineitem into corresponding Column's uniqueValues
func (model *Model) removeUniqueValues(lis ...LineItem) {
	for _, li := range lis {
		for _, column := range model.Columns {
			if value, ok := li.Get(column.Name); ok {
				column.RemoveUniquenessOf(value)
			}
		}
	}
}

//------End Columns Uniqueness------//

//------Check------//
// CheckRelationshipsMeta check the uniqueness of every element in HasOne and HasMany
func (model *Model) CheckRelationshipsMeta() error {
	set := gset.NewSetSimple()
	for _, resName := range model.HasMany {
		if set.Has(resName) {
			return HasManyErrorf("%s has been used", resName)
		} else {
			set.Add(resName)
		}
	}
	for _, resName := range model.HasOne {
		if set.Has(resName) {
			return HasOneErrorf("%s has been used", resName)
		} else {
			set.Add(resName)
		}
	}
	return nil
}

// checkColumnsMeta checks columns:
//   1. id must be the first column, its type must be number
//   2. CheckMeta
func (model *Model) CheckColumnsMeta() error {
	if len(model.Columns) < 1 ||
		model.Columns[0].Name != "id" ||
		model.Columns[0].Type != number.Name() {
		return ColumnsErrorf("The first colmun must be id with number type in file: %s", model.router.filePath)
	}

	for _, column := range model.Columns {
		if err := column.CheckMeta(); err != nil {
			return err
		}
	}

	return nil
}

// CheckRelationship
//   1. checks if every resource in HasOne and HasMany exists
//   2. CheckRelationships
func (model *Model) CheckRelationship(seed map[string]interface{}) error {
	for _, resoureName := range model.HasOne {
		if _, ok := model.router.apiFaker.Routers[inflection.Plural(resoureName)]; !ok {
			return HasOneErrorf("use unknown reource %s in file: %s", resoureName, model.router.filePath)
		}
	}
	for _, resoureName := range model.HasMany {
		if _, ok := model.router.apiFaker.Routers[inflection.Plural(resoureName)]; !ok {
			return HasManyErrorf("use unknown reource \"%s\" in file: %s", resoureName, model.router.filePath)
		}
	}

	for _, column := range model.Columns {
		if err := column.CheckRelationships(seed[column.Name], model); err != nil {
			return err
		}
	}
	return nil
}

// CheckRelationships
func (model *Model) CheckRelationships() error {
	for _, seed := range model.Seeds {
		if err := model.CheckRelationship(seed); err != nil {
			return err
		}
	}

	return nil
}

// ValidateValue checks specific seed
func (model *Model) ValidateValue(seed map[string]interface{}) error {
	columns := model.Columns

	if len(seed) != len(columns) {
		return SeedsErrorf("has wrong number of columns: %v", seed)
	}

	for _, column := range columns {
		if seedVal, ok := seed[column.Name]; !ok {
			return SeedsErrorf("has no column \"%s\" in seed: %v", column.Name, seed)
		} else {
			if err := column.CheckValue(seedVal, model); err != nil {
				return err
			}
		}
	}

	return nil
}

// ValidateSeedsValue
func (model *Model) ValidateSeedsValue() error {
	for _, seed := range model.Seeds {
		if err := model.ValidateValue(seed); err != nil {
			return err
		}
	}

	return nil
}

// Validate ValidateValue and CheckRelationships
func (model *Model) Validate(seed map[string]interface{}) error {
	return gtester.NewCheckQueue().Add(func() error {
		return model.ValidateValue(seed)
	}).Add(func() error {
		return model.CheckRelationship(seed)
	}).Run()
}

// CheckUniqueness check uniqueness for initialization for ApiFaker
func (model *Model) CheckUniqueness() error {
	// check id
	if !model.dataChanged && len(model.Seeds) != model.Len() {
		return SeedsErrorf("model[name=\"%s\"] has same id", model.Name)
	}

	// check other unique columns
	for _, column := range model.Columns {
		if column.Unique && column.uniqueValues.Len() != model.Len() {
			return SeedsErrorf("column[name=\"%s\"] in model[name=\"%s\"] has same values", column.Name, model.Name)
		}
	}

	return nil
}

//------End Check------//

//------Seeds and Set------//
// initSet adds all LineItem into Set, addUniqueValues and updateId
func (model *Model) initSet() {
	if model.Set == nil {
		model.Set = gset.NewSetThreadSafe()
	}
	for _, seed := range model.Seeds {
		li := NewLineItemWithMap(seed)
		model.Set.Add(li)
		model.addUniqueValues(li)
		model.updateId(li.ID())
	}
}

// backfillSeeds
func (model *Model) backfillSeeds() {
	model.RLock()
	defer model.RUnlock()

	models := model.ToLineItems()
	sort.Sort(models)
	model.Seeds = models.ToSlice()
	model.dataChanged = false
}

//------End Seeds and Set------//

// ToLineItems allocate a new LineItems filled with Model elements slice
func (model *Model) ToLineItems() LineItems {
	lis := []LineItem{}
	elements := model.Set.ToSlice()
	for _, element := range elements {
		if li, ok := element.(LineItem); ok {
			newLi := li.InsertRelatedData(model)
			lis = append(lis[:], newLi)
		}
	}
	return LineItems(lis)
}

// SaveToFile save model to file with the given path
func (model *Model) SaveToFile(path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	if model.dataChanged {
		model.backfillSeeds()
	}
	bytes, err := json.Marshal(model)
	if err != nil {
		return err
	}
	_, err = file.WriteString(string(bytes))

	return err
}
