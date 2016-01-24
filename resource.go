package apifaker

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Focinfi/gset"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
	"os"
)

var ColumnCountError = errors.New("Has wrong count of columns")
var ColumnNameError = errors.New("Has wrong column")

type Column struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type Item map[string]interface{}

func (i Item) Element() interface{} {
	return i["id"]
}

func (i *Item) UpdateWithGinContext(ctx *gin.Context, r *Resource) error {
	item := (map[string]interface{})(*i)
	for _, column := range r.Columns {
		if value := ctx.PostForm(column.Name); value != "" {
			// TODO: check type
			item[column.Name] = value
		}
	}
	tempItem := Item(item)
	i = &tempItem
	return nil
}

type Items []Item

func NewItemsFromInterfaces(elements []interface{}) Items {
	items := []Item{}
	for _, element := range elements {
		if item, ok := element.(Item); ok {
			items = append(items[:], item)
		}
	}
	return Items(items)
}

func (items Items) Len() int {
	return len(items)
}

func (items Items) Less(i, j int) bool {
	itemFirstId := items[i]["id"].(int)
	itemSecondId := items[j]["id"].(int)
	return itemFirstId < itemSecondId
}

func (items Items) Swap(i, j int) {
	tmpItem := items[i]
	items[i] = items[j]
	items[j] = tmpItem
}

type Resource struct {
	Name  string                   `json:"resource_name"`
	Seeds []map[string]interface{} `json:"seeds"`

	Columns []Column `json:"columns"`

	*gset.Set `json:"-"`
	filePath  string `json:"-"`
	nextId    int    `json:"-"`
}

func (r *Resource) Add(item map[string]interface{}) error {
	if err := r.checkSeed(item); err != nil {
		return err
	} else {
		item["id"] = r.nextId
		r.Set.Add(Item(item))
		r.nextId++
	}

	return nil
}

func (r *Resource) UpdateItem(id int, item *Item) error {
	itemTmp := (map[string]interface{})(*item)
	if err := r.checkSeed(itemTmp); err != nil {
		return err
	} else {
		itemTmp["id"] = id
		newItem := Item(itemTmp)
		r.Set.Add(newItem)
		item = &newItem
	}
	return nil
}

func (r *Resource) UpdateWithAttrsInGinContext(id int, ctx *gin.Context) (int, interface{}) {
	// check if element does exsit
	element, ok := r.Get(id)
	if !ok {
		return http.StatusNotFound, nil
	}

	// update item
	item := element.(Item)
	err := (&item).UpdateWithGinContext(ctx, r)
	if err != nil {
		return http.StatusBadRequest, map[string]string{"message": err.Error()}
	} else {
		return http.StatusOK, item
	}
}

func (r *Resource) UpdateWithAllAttrsInGinContex(id int, ctx *gin.Context) (int, interface{}) {
	// check if element does exsit
	_, ok := r.Get(id)
	if !ok {
		return http.StatusNotFound, nil
	}

	// create a new item
	newItem, err := NewItemWithGinContext(ctx, r)
	// update item
	r.UpdateItem(id, &newItem)

	if err != nil {
		return http.StatusBadRequest, map[string]string{"message": err.Error()}
	} else {
		return http.StatusOK, newItem
	}
}

// checkSeed check specific seed
func (r *Resource) checkSeed(seed map[string]interface{}) error {
	columns := r.Columns

	if len(seed) != len(columns) {
		return ColumnCountError
	}

	for _, column := range columns {
		if _, ok := seed[column.Name]; !ok {
			ColumnNameError = fmt.Errorf("has wrong column: %s", column.Name)
			return ColumnNameError
		}
		// TODO: check type
	}

	return nil
}

// checkSeeds() check if every item of this Resource object's Seeds
// is in line of description of its Columns.
func (r *Resource) checkSeeds() error {
	for _, seed := range r.Seeds {
		if err := r.checkSeed(seed); err != nil {
			return err
		}
	}

	return nil
}

// setItems
func (r *Resource) setItems() {
	if r.Set == nil {
		r.Set = gset.NewSet()
	}
	for _, seed := range r.Seeds {
		r.Add(Item(seed))
	}
}

// NewResourceWithPath allocates and returns a new Resource,
// using the given path as it's json file path
func NewResourceWithPath(path string) (r *Resource, err error) {
	// open file
	file, err := os.Open(path)
	if err != nil {
		return
	}
	defer file.Close()

	// read and unmarshal file
	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		return
	}
	r = &Resource{filePath: path, nextId: 1}
	if err = json.Unmarshal(bytes, r); err != nil {
		err = fmt.Errorf("[apifaker] json format error: %s, file: %s", err.Error(), path)
	} else {
		// check Seeds
		err = r.checkSeeds()

		// set items
		if err == nil {
			r.setItems()
		}
	}

	return
}

func NewItemWithGinContext(ctx *gin.Context, r *Resource) (Item, error) {
	item := make(map[string]interface{})
	for _, column := range r.Columns {
		if param := ctx.PostForm(column.Name); param != "" {
			// TODO: check type
			item[column.Name] = param
		} else {
			ColumnNameError = fmt.Errorf("doesn't has column: %s", column.Name)
			return nil, ColumnNameError
		}
	}

	return Item(item), nil
}
