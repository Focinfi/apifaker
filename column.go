package apifaker

import (
	"fmt"
	. "github.com/Focinfi/gset"
	"github.com/jinzhu/inflection"
	"reflect"
	"regexp"
	"strings"
)

var ColumnCountError = fmt.Errorf("Has wrong count of columns")
var ColumnNameError = fmt.Errorf("Has wrong column")
var ColumnTypeError = fmt.Errorf("Column type wrong")
var ColumnUniqueError = fmt.Errorf("Column already exists")

type JsonType string

const (
	boolean JsonType = "boolean"
	number  JsonType = "number"
	str     JsonType = "string"
	array   JsonType = "array"
	object  JsonType = "object"
)

// Name return JsonType string itself
func (j JsonType) Name() interface{} {
	return string(j)
}

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

// jsonTypes contains a list a supportted supportted
var jsonTypes = NewSetSimple(boolean, number, str, array, object)

type Column struct {
	Name          string         `json:"name"`
	Type          string         `json:"type"`
	Unique        bool           `json:"unique"`
	RegexpPattern string         `json:"regexp_pattern"`
	uniqueValues  *SetThreadSafe `json:"-"`
}

// CheckType check type if is in the jsonTypes
func (c Column) CheckMeta() error {
	if c.Name == "" {
		return fmt.Errorf("colmun[%#v] must has a name", c)
	}

	if c.Type == "" {
		return fmt.Errorf("column %s, must has a type", c.Name)
	}

	// type must be in jsonTypes
	if !jsonTypes.Has(JsonType(c.Type)) {
		return fmt.Errorf("don't support type: %s, supportted types %#v", c.Type, jsonTypes.ToSlice())
	}

	// check regexp format
	if c.RegexpPattern != "" {
		if _, err := regexp.Compile(c.RegexpPattern); err != nil {
			return fmt.Errorf("regexp pattern format[%s] is wrong, err is: %s", c.RegexpPattern, err.Error())
		}
	}

	return nil
}

// CheckRelationships
func (column *Column) CheckRelationships(seedVal interface{}, model *Model) error {
	// check foreign key
	if strings.HasSuffix(column.Name, "_id") {
		resName := strings.TrimSuffix(column.Name, "_id")
		resPluralName := inflection.Plural(resName)
		if router, ok := model.router.apiFaker.Routers[resPluralName]; !ok {
			return fmt.Errorf("has no model: %s", resPluralName)
		} else {
			var hasSeed bool
			for _, seed := range router.Model.Seeds {
				id, ok := seed["id"]
				if ok && id == seedVal {
					hasSeed = true
					break
				}
			}
			if !hasSeed {
				return fmt.Errorf("has no lineitem[id=%v] of model[%s]", seedVal, resPluralName)
			}
		}
	}

	return nil
}

// CheckValue
func (column *Column) CheckValue(seedVal interface{}, model *Model) error {
	// check type
	goType := JsonType(column.Type).GoType()
	seedType := reflect.TypeOf(seedVal).String()

	if seedType != goType {
		ColumnTypeError = fmt.Errorf("column[%s] type is wrong, expected %s, current is %s", column.Name, goType, 2)
		return ColumnTypeError
	}

	// check regexp pattern matching
	if column.RegexpPattern != "" && column.Type == str.Name() {
		matched, err := regexp.Match(column.RegexpPattern, []byte(seedVal.(string)))
		if err != nil {
			return fmt.Errorf("column regexp %s has format error", column.RegexpPattern)
		} else if !matched {
			return fmt.Errorf("colmun[%s]: %#v, doesn't match %s", column.Name, seedVal, column.RegexpPattern)
		}
	}

	// check uniqueness
	if !column.CheckUniqueOf(seedVal) {
		ColumnUniqueError = fmt.Errorf("cloumn[%s]: %v already exists", column.Name, seedVal)
		return ColumnUniqueError
	}

	return nil
}

// CheckUniqueOf check if the given value exists
func (c *Column) CheckUniqueOf(value interface{}) bool {
	if !c.Unique || c.Name == "id" {
		return true
	}

	if c.uniqueValues == nil {
		c.uniqueValues = NewSetThreadSafe()
		return true
	} else {
		return !c.uniqueValues.Has(T(value))
	}
}

// AddValue add the give value into the Column's uniqueValues
func (c *Column) AddValue(value interface{}) {
	if !c.Unique {
		return
	}

	if c.uniqueValues == nil {
		c.uniqueValues = NewSetThreadSafe(T(value))
	} else {
		c.uniqueValues.Add(T(value))
	}
}

// RemoveValue remove the given value from Column's uniqueValues
func (c *Column) RemoveValue(value interface{}) {
	if !c.Unique {
		return
	}

	if c.uniqueValues == nil {
		c.uniqueValues = NewSetThreadSafe()
	} else {
		c.uniqueValues.Remove(T(value))
	}
}
