package apifaker

import (
	"fmt"
	. "github.com/Focinfi/gset"
	"github.com/jinzhu/inflection"
	"reflect"
	"regexp"
	"strings"
)

type JsonType string

const (
	boolean JsonType = "boolean"
	number  JsonType = "number"
	str     JsonType = "string"
	array   JsonType = "array"
	object  JsonType = "object"
)

// Name returns JsonType string itself
func (j JsonType) Name() string {
	return string(j)
}

// GoType returns the corresponding type in golang
func (j JsonType) GoType() string {
	switch j {
	case boolean:
		return "bool"
	case number:
		return "float64"
	case str:
		return "string"
	case array:
		return "[]interface {}"
	case object:
		return "map[string]interface {}"
	}
	return "nil"
}

// jsonTypes contains a list a supportted json types
var jsonTypes = NewSetSimple(boolean, number, str, array, object)

type Column struct {
	Name          string `json:"name"`
	Type          string `json:"type"`
	Unique        bool   `json:"unique"`
	RegexpPattern string `json:"regexp_pattern"`
	uniqueValues  *SetThreadSafe
}

func (column *Column) getUniqueValues() *SetThreadSafe {
	if column.uniqueValues == nil {
		column.uniqueValues = NewSetThreadSafe()
	}
	return column.uniqueValues
}

// CheckType checks
//   1. Name and Type must be present
//   2. Type must in jsonTypes
//   3. RegexpPattern must valid
func (column Column) CheckMeta() error {
	if column.Name == "" {
		return ColumnsErrorf("colmun[content=%v] must has a name", column)
	}

	columnLogName := fmt.Sprintf("column[name=\"%s\"]", column.Name)
	if column.Type == "" {
		return ColumnsErrorf("%s must has a type", columnLogName)
	}

	if !jsonTypes.Has(JsonType(column.Type)) {
		return ColumnsErrorf("%s use unsupportted type: %s, all supportted types: %v", columnLogName, column.Type, jsonTypes.ToSlice())
	}

	if column.RegexpPattern != "" {
		if _, err := regexp.Compile(column.RegexpPattern); err != nil {
			return ColumnsErrorf("%s has wrong regexp pattern format: %s, error: %v", columnLogName, column.RegexpPattern)
		}
	}

	return nil
}

// CheckRelationships checks the if resource exists with the xxx_id
func (column *Column) CheckRelationships(seedVal interface{}, model *Model) error {
	if !strings.HasSuffix(column.Name, "_id") {
		return nil
	}

	columnLogName := fmt.Sprintf("column[name=\"%s\"]", column.Name)
	resName := strings.TrimSuffix(column.Name, "_id")
	resPluralName := inflection.Plural(resName)
	router, ok := model.router.apiFaker.Routers[resPluralName]
	if ok {
		for _, li := range router.Model.ToLineItems() {
			id, ok := li.Get("id")
			if ok && id == seedVal {
				return nil
			}
		}
	}

	return ColumnsErrorf("%s has no item[id=%v] of resource[resource_name=\"%s\"]", columnLogName, seedVal, resPluralName)
}

// CheckValue checks the value to insert database
//   1. type
//   2. regexp pattern matching
//   3. uniqueness if unique is true
func (column *Column) CheckValue(seedVal interface{}, model *Model) error {
	columnLogName := fmt.Sprintf("column[name=\"%s\"]", column.Name)
	goType := JsonType(column.Type).GoType()
	jsonType := JsonType(column.Type).Name()
	seedType := reflect.TypeOf(seedVal).String()
	if seedType != goType {
		return ColumnsErrorf("%s has wrong type, expect a %s, but use a %s", columnLogName, jsonType, seedType)
	}

	if column.RegexpPattern != "" && column.Type == str.Name() {
		matched, err := regexp.Match(column.RegexpPattern, []byte(seedVal.(string)))
		if err == nil && !matched {
			return fmt.Errorf("%s mismatch regexp format, value: %v, format: %s", columnLogName, seedVal, column.RegexpPattern)
		}
	}

	if !column.CheckUniquenessOf(seedVal) {
		return ColumnsErrorf("%s item value %v already exists", columnLogName, seedVal)
	}

	return nil
}

// CheckUniquenessOf checks if the given value exists
func (column *Column) CheckUniquenessOf(value interface{}) bool {
	if !column.Unique || column.Name == "id" {
		return true
	}

	return !column.getUniqueValues().Has(T(value))
}

// AddValue add the give value into the Column's uniqueValues
func (column *Column) AddUniquenessOf(value interface{}) {
	if !column.Unique {
		return
	}

	column.getUniqueValues().Add(T(value))
}

// RemoveValue remove the given value from Column's uniqueValues
func (column *Column) RemoveUniquenessOf(value interface{}) {
	if !column.Unique {
		return
	}

	column.getUniqueValues().Remove(T(value))
}
