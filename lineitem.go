package apifaker

import (
	"fmt"
	"github.com/gin-gonic/gin"
)

type LineItem struct {
	dataMap map[string]interface{}
}

func NewLineItemWithMap(dataMap map[string]interface{}) LineItem {
	return LineItem{dataMap}
}

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

func (li LineItem) Len() int {
	return len(li.dataMap)
}

func (li LineItem) Get(key string) (interface{}, bool) {
	value, ok := li.dataMap[key]
	return value, ok
}

func (li *LineItem) Set(key string, value interface{}) {
	li.dataMap[key] = value
}

func (li LineItem) ToMap() map[string]interface{} {
	return li.dataMap
}

type LineItems []LineItem

func NewLineItemsFromInterfaces(elements []interface{}) LineItems {
	lis := []LineItem{}
	for _, element := range elements {
		if li, ok := element.(LineItem); ok {
			lis = append(lis[:], li)
		}
	}
	return LineItems(lis)
}

func (lis LineItems) Len() int {
	return len(lis)
}

func (lis LineItems) Less(i, j int) bool {
	firstId, _ := lis[i].Get("id")
	secondId, _ := lis[j].Get("id")
	return firstId.(int) < secondId.(int)
}

func (lis LineItems) Swap(i, j int) {
	tmpItem := lis[i]
	lis[i] = lis[j]
	lis[j] = tmpItem
}

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

func (lis LineItems) ToSlice() []map[string]interface{} {
	slice := []map[string]interface{}{}
	for _, li := range lis {
		slice = append(slice, li.ToMap())
	}
	return slice
}
