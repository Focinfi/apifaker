package apifaker

import (
	"github.com/Focinfi/gtester"
	"testing"
)

func TestNewModelWithPath(t *testing.T) {
	model, err := NewModelWithPath(testDir + "/users.json")
	if err != nil {
		t.Error(err)
	}
	gtester.AssertEqual(t, model.Name, "users")
	gtester.AssertEqual(t, model.Set.Len(), 3)
}

func TestCheckModelColumns(t *testing.T) {
	// has wrong columns count
	model, _ := NewModelWithPath(testDir + "/users.json")
	model.Seeds = append(model.Seeds, map[string]interface{}{})
	if err := model.checkSeeds(); err != ColumnCountError {
		t.Error("Can not detect columns count of seed item")
	}

	// has wrong column
	model, _ = NewModelWithPath(testDir + "/users.json")
	model.Seeds[0] = map[string]interface{}{"foo1": "bar", "foo2": "bar", "foo3": 1}
	if err := model.checkSeeds(); err != ColumnNameError {
		t.Error("Can not check column name")
	}
}
