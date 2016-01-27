package apifaker

import (
	"fmt"
	. "github.com/Focinfi/gset"
)

var ColumnCountError = fmt.Errorf("Has wrong count of columns")
var ColumnNameError = fmt.Errorf("Has wrong column")
var ColumnTypeError = fmt.Errorf("Column type wrong")

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
	Name   string `json:"name"`
	Type   string `json:"type"`
	Unique bool   `json:"unique"`
	Format string `json:"format_regex"`
}

// CheckType check type if is in the jsonTypes
func (c Column) CheckTypeMeta() error {
	if c.Name == "" {
		return fmt.Errorf("colmun[%#v] must has a name", c)
	}

	if c.Type == "" {
		return fmt.Errorf("column %s, must has a type", c.Name)
	}

	if !jsonTypes.Has(T(c.Type)) {
		return fmt.Errorf("don't support type: %s, supportted types %#v", jsonTypes.ToSlice())
	}
	return nil
}
