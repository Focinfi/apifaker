package apifaker

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Focinfi/gset"
	// "github.com/gin-gonic/gin"
	"io/ioutil"
	"os"
)

var CloumnCountError = errors.New("Has wrong count of columns")
var CloumnNameError = errors.New("Has wrong column")

type Cloumn struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type Item map[string]interface{}

func (i Item) Element() interface{} {
	return i["id"]
}

type Resource struct {
	Name  string                   `json:"resource_name"`
	Seeds []map[string]interface{} `json:"seeds"`

	Cloumns []Cloumn `json:"cloumns"`

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

// checkSeed check specific seed
func (r *Resource) checkSeed(seed map[string]interface{}) error {
	columns := r.Cloumns

	if len(seed) != len(columns) {
		return CloumnCountError
	}

	for _, column := range columns {
		if _, ok := seed[column.Name]; !ok {
			return CloumnNameError
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

// func NewResourceWithGinContext(ctx *gin.Context) (*Resource, error) {

// }
