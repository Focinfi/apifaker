package apifaker

import (
	. "github.com/Focinfi/gtester"
	"os"
	"testing"
)

var testRouter = &Router{apiFaker: &ApiFaker{}}

func TestNewModelWithPath(t *testing.T) {
	model, err := NewModelWithPath(testDir+"/users.json", testRouter)
	if err != nil {
		t.Error(err)
	}

	AssertEqual(t, model.Name, "users")
	AssertEqual(t, model.HasMany, []string{"books"})
}

func TestCheckRelationships(t *testing.T) {
	model, _ := NewModelWithPath(testDir+"/users.json", testRouter)
	model.HasMany = append(model.HasMany, model.HasMany...)
	AssertError(t, model.CheckRelationshipsMeta())
}

func TestCheckModelColumns(t *testing.T) {
	// has wrong columns count
	model, _ := NewModelWithPath(testDir+"/users.json", testRouter)
	model.Seeds[0] = map[string]interface{}{}
	if err := model.checkSeeds(); err == nil {
		t.Errorf("Can not detect columns count of seed item")
	}

	// has wrong column
	model, _ = NewModelWithPath(testDir+"/users.json", testRouter)
	model.Seeds[0] = map[string]interface{}{"foo0": "bar", "foo1": "bar", "foo2": "bar", "foo3": 1}
	if err := model.checkSeeds(); err == nil {
		t.Errorf("Can not check column name")
	}
}
func TestCheckColumnsTypes(t *testing.T) {
	model, _ := NewModelWithPath(testDir+"/users.json", testRouter)
	model.Columns[0].Type = "unsupportted_type"
	AssertError(t, model.checkColumnsMeta())
}

func TestCheckSeed(t *testing.T) {
	faker, _ := NewWithApiDir(testDir)
	userModel := faker.Routers["users"].Model
	bookModel := faker.Routers["books"].Model
	// type
	userModel.Seeds[0]["age"] = "22"
	AssertError(t, userModel.checkSeed(userModel.Seeds[0]))

	// regexp
	userModel.Seeds[1]["name"] = ",,,"
	AssertError(t, userModel.checkSeed(userModel.Seeds[1]))

	// uniqueness
	AssertError(t, userModel.checkSeed(userModel.Seeds[2]))

	// reset after delete
	firstLi, _ := userModel.Get(0)
	toDeleteName, _ := firstLi.Get("name")
	userModel.Delete(0)
	AssertEqual(t, userModel.Columns[0].CheckUniquenessOf(toDeleteName), true)

	// relationship
	book := map[string]interface{}{"id": float64(4), "title": "Animal Farm", "user_id": float64(100)}
	AssertError(t, bookModel.checkSeed(book))
	// t.Log(bookModel.checkSeed(book))
}

func TestSaveToFile(t *testing.T) {
	model, _ := NewModelWithPath(testDir+"/users.json", testRouter)
	liMap := map[string]interface{}{
		"id":    float64(4),
		"name":  "Monica",
		"phone": "12332132132",
		"age":   float64(21),
	}

	li := LineItem{liMap}
	model.Add(li)
	model.backfillSeeds()
	AssertEqual(t, len(model.Seeds), 4)
	AssertEqual(t, model.dataChanged, false)

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
