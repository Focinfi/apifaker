package apifaker

import (
	. "github.com/smartystreets/goconvey/convey"
	"os"
	"testing"
)

var testRouter = &Router{apiFaker: &ApiFaker{}}

func validModel(name string) *Model {
	faker, _ := NewWithApiDir(testDir)
	return faker.Routers[name].Model
}

func validUserModel() *Model {
	return validModel("users")
}

func validBookModel() *Model {
	return validModel("books")
}

func TestModel(t *testing.T) {
	model, err := NewModelWithPath(testDir+"/users.json", testRouter)
	Describ("NewModelWithPath", t, func() {
		It("allocates and returns a new Model", func() {
			Expect(err, ShouldBeNil)
			Expect(model.currentId, ShouldEqual, 3)
			Expect(model.dataChanged, ShouldBeFalse)
		})
	})

	Describ("InsertRelatedData", t, func() {
		model := validUserModel()
		li, _ := model.Get(float64(1))
		newLi, _ := li.InsertRelatedData(model)
		_, hasbooks := newLi.Get("books")
		_, hasAvatar := newLi.Get("avatar")
		It("inserts books and avatar columns into li", func() {
			Expect(newLi.Len(), ShouldEqual, 6)
			Expect(hasbooks, ShouldBeTrue)
			Expect(hasAvatar, ShouldBeTrue)
		})
	})

	Describ("CheckRelationshipsMeta", t, func() {
		model := validUserModel()
		model.HasMany = append(model.HasMany, model.HasMany...)
		It("returns error if has_many has repeated elements", func() {
			Expect(model.CheckRelationshipsMeta(), ShouldNotBeNil)
		})
	})

	Describ("checkColumnsMeta", t, func() {
		Context("when first item is not id", func() {
			It("returns error", func() {
				model := validUserModel()
				model.Columns[0] = &Column{Name: "idea"}
				Expect(model.CheckColumnsMeta(), ShouldNotBeNil)
			})
		})
		Context("when type is nil", func() {
			It("returns error", func() {
				model := validUserModel()
				model.Columns[0] = &Column{Name: "id"}
				Expect(model.CheckColumnsMeta(), ShouldNotBeNil)
			})
		})
		Context("when type is not supportted", func() {
			It("returns error", func() {
				model := validUserModel()
				model.Columns[0] = &Column{Name: "id", Type: "xxx"}
				Expect(model.CheckColumnsMeta(), ShouldNotBeNil)
			})
		})
		Context("when regexp format is wrong", func() {
			It("returns error", func() {
				model := validUserModel()
				model.Columns[0] = &Column{Name: "id", Type: "number", RegexpPattern: "(?=)"}
				Expect(model.CheckColumnsMeta(), ShouldNotBeNil)
			})
		})
	})

	Describ("checkSeeds", t, func() {
		Context("when has wrong columns count", func() {
			It("returns error", func() {
				model := validUserModel()
				delete(model.Seeds[0], "id")
				Expect(model.CheckSeeds(), ShouldNotBeNil)
			})
		})
		Context("when has wrong colum", func() {
			It("returns error", func() {
				model := validUserModel()
				delete(model.Seeds[0], "id")
				model.Seeds[0]["xxx"] = true
				Expect(model.CheckSeeds(), ShouldNotBeNil)
			})
		})
		Context("when type is wrong", func() {
			It("returns error", func() {
				model := validUserModel()
				model.Seeds[0]["name"] = 1010
				Expect(model.CheckSeeds(), ShouldNotBeNil)
			})
		})
		Context("when doesn't much regexp pattern", func() {
			It("returns error", func() {
				model := validUserModel()
				model.Seeds[0]["name"] = "~.~"
				Expect(model.CheckSeeds(), ShouldNotBeNil)
			})
		})
		Context("when column value has been existed", func() {
			It("returns error", func() {
				model := validUserModel()
				model.Seeds[0]["name"] = model.Seeds[1]["name"]
				Expect(model.CheckSeeds(), ShouldNotBeNil)
			})
		})
		Context("when column has a inexistent foreign key", func() {
			It("returns error", func() {
				model := validBookModel()
				model.Seeds[0]["title"] = "On the way"
				model.Seeds[0]["user_id"] = float64(100)
				Expect(model.CheckSeeds(), ShouldNotBeNil)
			})
		})
	})

	Describ("SaveToFile", t, func() {
		model := validUserModel()
		model.Add(LineItem{map[string]interface{}{
			"id":    float64(4),
			"name":  "Monica",
			"phone": "12332132132",
			"age":   21,
		}})
		model.backfillSeeds()
		It("set dataChange to be false", func() {
			Expect(model.dataChanged, ShouldBeFalse)
		})

		err := model.SaveToFile(testDir + "/users_temp.json")
		defer os.Remove(testDir + "/users_temp.json")
		It("saves to file", func() {
			Expect(err, ShouldBeNil)
		})
	})
}
