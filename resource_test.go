package apifaker

import (
	"github.com/Focinfi/gtester"
	"testing"
)

func TestNewResourceWithPath(t *testing.T) {
	resource, err := NewResourceWithPath(testDir + "/users.json")
	if err != nil {
		t.Error(err)
	}
	gtester.AssertEqual(t, resource.Name, "users")
	gtester.AssertEqual(t, resource.Set.Len(), 3)
}

func TestCheckColumns(t *testing.T) {
	// has wrong columns count
	resource, _ := NewResourceWithPath(testDir + "/users.json")
	resource.Seeds = append(resource.Seeds, map[string]interface{}{})
	if err := resource.checkSeeds(); err != ColumnCountError {
		t.Error("Can not detect columns count of seed item")
	}

	// has wrong column
	resource, _ = NewResourceWithPath(testDir + "/users.json")
	resource.Seeds[0] = map[string]interface{}{"foo1": "bar", "foo2": "bar", "foo3": 1}
	if err := resource.checkSeeds(); err != ColumnNameError {
		t.Error("Can not check column name")
	}
}
