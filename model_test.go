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
		newLi := li.InsertRelatedData(model)
		_, hasbooks := newLi.Get("books")
		_, hasAvatar := newLi.Get("avatar")
		It("inserts books and avatar columns into li", func() {
			Expect(newLi.Len(), ShouldEqual, 6)
			Expect(hasbooks, ShouldBeTrue)
			Expect(hasAvatar, ShouldBeTrue)
		})
	})

	Describ("Uniqueness", t, func() {
		Context("when check a user name already exists", func() {
			model := validBookModel()
			li, _ := model.Get(float64(1))
			titleColumn := model.Columns[1]
			name, _ := li.Get("title")
			It("returns error", func() {
				Expect(titleColumn.CheckValue(name, model), ShouldNotBeNil)
			})
		})

		Context("when check a title which has been edited", func() {
			model := validBookModel()
			li, _ := model.Get(float64(1))
			titleColumn := model.Columns[1]
			oldName, _ := li.Get("title")
			newLi := NewLineItemWithMap(li.ToMap())
			newLi.Set("title", "XXX")
			model.Update(float64(1), &newLi)
			It("returns nil error", func() {
				Expect(titleColumn.CheckValue(oldName, model), ShouldBeNil)
			})
		})
	})

	Describ("CheckRelationshipsMeta", t, func() {
		Context("when has_many has repeated elements", func() {
			model := validUserModel()
			model.HasMany = append(model.HasMany, model.HasMany...)
			It("returns error ", func() {
				Expect(model.CheckRelationshipsMeta(), ShouldNotBeNil)
			})
		})
		Context("when has_one has repeated elements", func() {
			model := validUserModel()
			model.HasOne = append(model.HasOne, model.HasMany...)
			It("returns error ", func() {
				Expect(model.CheckRelationshipsMeta(), ShouldNotBeNil)
			})
		})
	})

	Describ("CheckRelationships", t, func() {
		Context("when has_one or has_many has unknown resources", func() {
			model := validUserModel()
			model.HasMany = append(model.HasMany, "foo")
			It("returns error ", func() {
				Expect(model.CheckRelationships(), ShouldNotBeNil)
			})
		})
		Context("when has_one or has_one has unknown resources", func() {
			model := validUserModel()
			model.HasOne = append(model.HasOne, "foo")
			It("returns error ", func() {
				Expect(model.CheckRelationships(), ShouldNotBeNil)
			})
		})
	})

	Describ("CheckColumnsMeta", t, func() {
		Context("when first item is not id", func() {
			It("returns error", func() {
				model := validUserModel()
				model.Columns[0] = &Column{Name: "idea"}
				Expect(model.CheckColumnsMeta(), ShouldNotBeNil)
			})
		})
		Context("when id's type is nil", func() {
			It("returns error", func() {
				model := validUserModel()
				model.Columns[0] = &Column{Name: "id"}
				Expect(model.CheckColumnsMeta(), ShouldNotBeNil)
			})
		})
		Context("when name is nil", func() {
			It("returns error", func() {
				model := validUserModel()
				model.Columns[1].Name = ""
				Expect(model.CheckColumnsMeta(), ShouldNotBeNil)
			})
		})
		Context("when type is nil", func() {
			It("returns error", func() {
				model := validUserModel()
				model.Columns[1].Type = ""
				Expect(model.CheckColumnsMeta(), ShouldNotBeNil)
			})
		})
		Context("when type is not in jsonTypes", func() {
			It("returns error", func() {
				model := validUserModel()
				model.Columns[1].Type = "xxx"
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

	Describ("ValidateSeedsValue", t, func() {
		Context("when has wrong columns count", func() {
			It("returns error", func() {
				model := validUserModel()
				delete(model.Seeds[0], "id")
				Expect(model.ValidateSeedsValue(), ShouldNotBeNil)
			})
		})
		Context("when has wrong colum", func() {
			It("returns error", func() {
				model := validUserModel()
				delete(model.Seeds[0], "id")
				model.Seeds[0]["xxx"] = true
				Expect(model.ValidateSeedsValue(), ShouldNotBeNil)
			})
		})
		Context("when type is wrong", func() {
			It("returns error", func() {
				model := validUserModel()
				model.Seeds[0]["name"] = 1010
				Expect(model.ValidateSeedsValue(), ShouldNotBeNil)
			})
		})
		Context("when doesn't much regexp pattern", func() {
			It("returns error", func() {
				model := validUserModel()
				model.Seeds[0]["name"] = "~.~"
				Expect(model.ValidateSeedsValue(), ShouldNotBeNil)
			})
		})
		Context("when column value has been existed", func() {
			It("returns error", func() {
				model := validUserModel()
				model.Seeds[0]["name"] = model.Seeds[1]["name"]
				Expect(model.ValidateSeedsValue(), ShouldNotBeNil)
			})
		})
		Context("when column has a inexistent foreign key", func() {
			It("returns error", func() {
				model := validBookModel()
				model.Seeds[0]["title"] = "On the way"
				model.Seeds[0]["user_id"] = float64(100)
				Expect(model.Validate(model.Seeds[0]), ShouldNotBeNil)
			})
		})
	})

	Describ("SaveToFile", t, func() {
		model := validUserModel()
		err := model.Add(LineItem{map[string]interface{}{
			"id":    float64(4),
			"name":  "Monica",
			"phone": "12332132132",
			"age":   float64(21),
		}})

		It("set dataChanged to be true", func() {
			Expect(model.dataChanged, ShouldBeTrue)
		})

		err = model.SaveToFile(testDir + "/users_temp.json.test")
		defer os.Remove(testDir + "/users_temp.json.test")
		It("saves to file", func() {
			Expect(err, ShouldBeNil)
		})
	})
}
