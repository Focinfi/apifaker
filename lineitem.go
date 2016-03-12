package apifaker

import (
	"fmt"
	"github.com/gin-gonic/gin"

	"github.com/jinzhu/inflection"
	"sort"
	"strconv"
)

type LineItem struct {
	dataMap map[string]interface{}
}

// NewLineItemWithMap allocates and returns a new LineItem,
// using the dataMap param as the new LineItem's dataMap
func NewLineItemWithMap(dataMap map[string]interface{}) LineItem {
	return LineItem{dataMap}
}

// NewLineItemWithGinContext allocates and returns a new LineItem,
// its keys are from Model.Cloumns, values are from gin.Contex.PostForm(),
// error will be not nil if gin.Contex.PostForm() has no value for any key
func NewLineItemWithGinContext(ctx *gin.Context, model *Model) (LineItem, error) {
	li := LineItem{make(map[string]interface{})}
	for _, column := range model.Columns {
		// skip id column
		if column.Name == "id" {
			continue
		}
		if value := ctx.PostForm(column.Name); value != "" {
			li.SetStringValue(column.Name, value, column.Type)
		} else {
			return li, fmt.Errorf("doesn't has column: %s", column.Name)
		}
	}

	return li, nil
}

// ID returns the float64 of id
func (li *LineItem) ID() float64 {
	return li.Id().(float64)
}

// Id returns the value of LineItem's dataMap["id"]
// it will panic if got a nil or a not float64 "id"
func (li LineItem) Id() interface{} {
	if id, ok := li.Get("id"); ok {
		if idFloat64, ok := id.(float64); ok {
			return idFloat64
		} else {
			panic(fmt.Sprintf("[LineItem]: id must be a float64, %#v\n", li.ToMap()))
		}
	} else {
		panic(fmt.Sprintf("[LineItem]: must has id: %#v\n", li.ToMap()))
	}
}

// Len returns LineItem's dataMap's length
func (li LineItem) Len() int {
	return len(li.dataMap)
}

// Get gets dataMap[key] and returns its value and exsiting in LinItem
func (li LineItem) Get(key string) (interface{}, bool) {
	value, ok := li.dataMap[key]
	return value, ok
}

// Set set key-value of dataMap in LineItem
func (li *LineItem) Set(key string, value interface{}) {
	li.dataMap[key] = value
}

// SetStringValue formats the given string value with given valueType,
// sets key into LineItem formated value
func (li *LineItem) SetStringValue(key, value, valueType string) error {
	formatValue, err := FormatValue(valueType, value)
	if err != nil {
		return err
	} else {
		li.Set(key, formatValue)
	}
	return nil
}

// InsertRelatedData allocates and returns a new LineItem,
// it will has all data of the caller LineItem,
// it will insert all related data if the given Model's has any Column named xxx_id
func (li LineItem) InsertRelatedData(model *Model) LineItem {
	// has one relationship
	newLi := NewLineItemWithMap(li.ToMap())
	singularName := inflection.Singular(model.Name)
	// HasOne
	for _, resName := range model.HasOne {
		resStruct := map[string]interface{}{}
		if resRouter, ok := model.router.apiFaker.Routers[inflection.Plural(resName)]; ok {
			resLis := resRouter.Model.ToLineItems()
			sort.Sort(resLis)
			for _, resLi := range resLis {
				curId, ok := resLi.Get(fmt.Sprintf("%s_id", singularName))
				if ok && curId == newLi.Id() {
					resStruct = resLi.ToMap()
					break
				}
			}
		}
		if len(resStruct) > 0 {
			newLi.Set(inflection.Singular(resName), resStruct)
		}
	}

	// HasMany
	for _, resName := range model.HasMany {
		resSlice := []interface{}{}
		if resRouter, ok := model.router.apiFaker.Routers[resName]; ok {
			resLis := resRouter.Model.ToLineItems()
			sort.Sort(resLis)
			for _, resLi := range resLis {
				curId, ok := resLi.Get(fmt.Sprintf("%s_id", singularName))
				if ok && curId == newLi.Id() {
					resSlice = append(resSlice, resLi.ToMap())
				}
			}
		}
		if len(resSlice) > 0 {
			newLi.Set(resName, resSlice)
		}
	}

	return newLi
}

// DeleteRelatedLis deletes all related data
func (li LineItem) DeleteRelatedLis(id float64, model *Model) {
	_, ok := model.Get(id)
	if !ok {
		return
	}

	for _, rotuer := range model.router.apiFaker.Routers {
		var isRelatedRouter bool
		var foreign_key = fmt.Sprintf("%s_id", inflection.Singular(model.Name))
		for _, column := range rotuer.Model.Columns {
			if column.Name == foreign_key {
				isRelatedRouter = true
				break
			}
		}

		if isRelatedRouter {
			for _, li := range rotuer.Model.ToLineItems() {
				key, ok := li.Get(foreign_key)
				if ok && key == id {
					rotuer.Model.Delete(li.ID())
				}
			}
		}
	}
}

// FormatValue format the given string value described by the given valueType
func FormatValue(valueType, value string) (interface{}, error) {
	var nilValue interface{}
	switch valueType {
	case number.Name():
		if numberVal, err := strconv.ParseFloat(value, 64); err != nil {
			return nilValue, err
		} else {
			return numberVal, nil
		}
	case boolean.Name():
		if booleanVal, err := strconv.ParseBool(value); err != nil {
			return nilValue, err
		} else {
			return booleanVal, nil
		}
	default:
		return value, nil
	}
}

// ToMap allocates and returns a new map[string]interface{} filled with LineItem's dataMap
func (li LineItem) ToMap() map[string]interface{} {
	newMap := map[string]interface{}{}
	for k, v := range li.dataMap {
		newMap[k] = v
	}
	return newMap
}

type LineItems []LineItem

// Len returns LineItems's length
func (lis LineItems) Len() int {
	return len(lis)
}

// Len returns comparation of value's of two LineItem's "id"
func (lis LineItems) Less(i, j int) bool {
	return lis[i].ID() < lis[j].ID()
}

// Swap swaps two LineItem
func (lis LineItems) Swap(i, j int) {
	tmpItem := lis[i]
	lis[i] = lis[j]
	lis[j] = tmpItem
}

// ToSlice return a []map[string]interface{} filled with LineItems' elements
func (lis LineItems) ToSlice() []map[string]interface{} {
	slice := []map[string]interface{}{}
	for _, li := range lis {
		slice = append(slice, li.ToMap())
	}
	return slice
}
