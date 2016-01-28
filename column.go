package apifaker

import (
	"fmt"
	. "github.com/Focinfi/gset"
	"reflect"
	"regexp"
)

var ColumnCountError = fmt.Errorf("Has wrong count of columns")
var ColumnNameError = fmt.Errorf("Has wrong column")
var ColumnTypeError = fmt.Errorf("Column type wrong")
var ColumnUniqueError = fmt.Errorf("Column already exists")

type JsonType string

func (j JsonType) Element() interface{} {
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
		return "[]map[string]interface{}"
	case object:
		return "map[string]interface{}"
	}
	return "nil"
}

const (
	boolean JsonType = "boolean"
	number  JsonType = "number"
	str     JsonType = "string"
	array   JsonType = "array"
	object  JsonType = "object"
)

// jsonTypes contains a list a supportted supportted
var jsonTypes = NewSet(boolean, number, str, array, object)

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
	if !jsonTypes.Has(T(c.Type)) {
		return fmt.Errorf("don't support type: %s, supportted types %#v", jsonTypes.ToSlice())
	}

	// check regexp format
	if c.RegexpPattern != "" {
		if _, err := regexp.Compile(c.RegexpPattern); err != nil {
			return fmt.Errorf("regexp pattern format[%s] is wrong, err is: %s", c.RegexpPattern, err.Error())
		}
	}

	return nil
}

// CheckValue check value if valid
func (column *Column) CheckValue(seedVal interface{}) error {
	// check type
	typeElement, _ := jsonTypes.Get(column.Type)
	goType := typeElement.(JsonType).GoType()
	seedType := reflect.TypeOf(seedVal).String()
	if seedType != goType {
		ColumnTypeError = fmt.Errorf("column[%s] type is wrong, expected %s, current is %s", column.Name, goType, seedType)
		return ColumnTypeError
	}

	// check regexp pattern matching
	if column.RegexpPattern != "" && column.Type == str.Element() {
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
	if !c.Unique {
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
