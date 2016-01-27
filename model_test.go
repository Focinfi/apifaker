package apifaker

import (
	. "github.com/Focinfi/gtester"
	"os"
	"testing"
)

func TestNewModelWithPath(t *testing.T) {
	model, err := NewModelWithPath(testDir + "/users.json")
	if err != nil {
		t.Error(err)
	}

	AssertEqual(t, model.Name, "users")
	AssertEqual(t, model.Set.Len(), 3)
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

func TestSaveToFile(t *testing.T) {
	model, _ := NewModelWithPath(testDir + "/users.json")
	liMap := map[string]interface{}{
		"name":  "Monica",
		"phone": "12332132132",
		"age":   float64(21),
	}

	li := LineItem{liMap}
	model.Add(li)
	model.SetToSeeds()
	AssertEqual(t, model.Has(4), true)
	AssertEqual(t, model.dataChanged, true)

	liInSeed, _ := model.Get(4)
	nameInSeed, _ := liInSeed.Get("name")
	AssertEqual(t, nameInSeed, liMap["name"])
	phoneInSeed, _ := liInSeed.Get("phone")
	AssertEqual(t, phoneInSeed, liMap["phone"])
	ageInSeed, _ := liInSeed.Get("age")
	AssertEqual(t, ageInSeed, liMap["age"])

	err := model.SaveToFile(testDir + "/users_temp.json")
	defer os.Remove(testDir + "/users_temp.json")

	if err != nil {
		t.Error(err.Error())
	} else {
		AssertEqual(t, model.dataChanged, false)
	}
}

func TestCheckColumnsTypes(t *testing.T) {
	model, _ := NewModelWithPath(testDir + "/users.json")
	model.Columns[0].Type = "unsupportted_type"
	AssertError(t, model.checkColumnsType())
}

func TestCheckSeed(t *testing.T) {
	model, _ := NewModelWithPath(testDir + "/users.json")
	model.Seeds[0]["age"] = "22"
	AssertError(t, model.checkSeed(model.Seeds[0]))
}
