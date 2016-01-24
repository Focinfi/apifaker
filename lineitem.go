package apifaker

import (
	"fmt"
	"github.com/gin-gonic/gin"
)

type LineItem struct {
	dataMap map[string]interface{}
}

// NewLineItemWithMap allocates and returns a new LineItem,
// using the dataMap param as the new LineItem's dataMap
func NewLineItemWithMap(dataMap map[string]interface{}) LineItem {
	return LineItem{dataMap}
}

// NewLineItemWithGinContext allocates and return a new LineItem,
// its keys are from Model.Cloumns, values ara from gin.Contex.PostForm(),
// error will be not nil if gin.Contex.PostForm() has not value for any key
func NewLineItemWithGinContext(ctx *gin.Context, r *Model) (LineItem, error) {
	li := LineItem{make(map[string]interface{})}
	for _, column := range r.Columns {
		if value := ctx.PostForm(column.Name); value != "" {
			// TODO: check type
			li.Set(column.Name, value)
		} else {
			ColumnNameError = fmt.Errorf("doesn't has column: %s", column.Name)
			return li, ColumnNameError
		}
	}

	return li, nil
}

// Element returns the value of LineItem's dataMap["id"]
// it will panic if got a nil or no-int "id"
func (li LineItem) Element() interface{} {
	if id, ok := li.Get("id"); ok {
		if idInt, ok := id.(int); ok {
			return idInt
		} else {
			panic(fmt.Sprintf("lineitem: %#v, its is id must be a int", li))
		}
	} else {
		panic(fmt.Sprintf("this lineitem has no id: %#v", li))
	}
}

// Len returns LineItem's dataMap's length
func (li LineItem) Len() int {
	return len(li.dataMap)
}

// Get get dataMap[key] and returns its value and exsiting in LinItem
func (li LineItem) Get(key string) (interface{}, bool) {
	value, ok := li.dataMap[key]
	return value, ok
}

// Set set key-value of dataMap in LineItem
func (li *LineItem) Set(key string, value interface{}) {
	li.dataMap[key] = value
}

// ToMap returns dataMap in LineItem
func (li LineItem) ToMap() map[string]interface{} {
	return li.dataMap
}

type LineItems []LineItem

// NewLineItemsFromInterfaces allocates and returns a new LineItems
// using elements param as its content
func NewLineItemsFromInterfaces(elements []interface{}) LineItems {
	lis := []LineItem{}
	for _, element := range elements {
		if li, ok := element.(LineItem); ok {
			lis = append(lis[:], li)
		}
	}
	return LineItems(lis)
}

// Len returns LineItems's length
func (lis LineItems) Len() int {
	return len(lis)
}

// Len returns comparation of value's of two LineItem's "id"
func (lis LineItems) Less(i, j int) bool {
	firstId, _ := lis[i].Get("id")
	secondId, _ := lis[j].Get("id")
	return firstId.(int) < secondId.(int)
}

// Swap swap two LineItem
func (lis LineItems) Swap(i, j int) {
	tmpItem := lis[i]
	lis[i] = lis[j]
	lis[j] = tmpItem
}

// ToSlice return a []map[string]interface{} filled with LineItems' content
func (lis LineItems) ToSlice() []map[string]interface{} {
	slice := []map[string]interface{}{}
	for _, li := range lis {
		slice = append(slice, li.ToMap())
	}
	return slice
}
