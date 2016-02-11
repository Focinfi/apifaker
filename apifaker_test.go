package apifaker

import (
	"fmt"
	"github.com/Focinfi/gtester"
	"github.com/Focinfi/gtester/httpmock"
	. "github.com/smartystreets/goconvey/convey"
	"net/http"
	"os"
	"testing"
)

var Describ = Convey
var Context = Convey
var It = So

var testDir = os.Getenv("GOPATH") + "/src/github.com/Focinfi/apifaker/api_static_test"

var userParam = map[string]interface{}{
	"name": "Ameng", "phone": "13213213214", "age": float64(22),
}

var invalidUserParam = map[string]interface{}{
	"name": "N", "phone": "13013213214", "age": "22x",
}

var usersFixture = []map[string]interface{}{
	{
		"id":    1,
		"name":  "Frank",
		"phone": "13213213213",
		"age":   float64(22),
		"books": []interface{}{
			map[string]interface{}{
				"id":      float64(1),
				"title":   "The Little Prince",
				"user_id": float64(1),
			},
			map[string]interface{}{
				"id":      float64(3),
				"title":   "The Alchemist",
				"user_id": float64(1),
			},
		},
		"avatar": map[string]interface{}{
			"id":      float64(1),
			"url":     "http://example.com/avatar.png",
			"user_id": float64(1),
		},
	},
	{"id": 2, "name": "Antony", "phone": "13213213211", "age": float64(22),
		"books": []interface{}{
			map[string]interface{}{
				"id":      float64(2),
				"title":   "Life of Pi",
				"user_id": float64(2),
			},
		},
	},
	{"id": 3, "name": "Foci", "phone": "13213213212", "age": float64(22)},
}

func shouldHasJsonResponse(response interface{}, exp ...interface{}) string {
	if record, ok := response.(*httpmock.Recorder); ok {
		if !gtester.ResponseEqual(record, exp[0]) {
			return fmt.Sprintf("Expect: '%v'\nActual: '%v\n", exp[0], record.Body)
		}
	} else {
		panic("First parameter of shouldHasJsonResponse() must be '*httpmock.Recorder'")
	}
	return ""
}

func TestApiFaker(t *testing.T) {
	faker, err := NewWithApiDir(testDir)
	Describ("NewWithDir", t, func() {
		It(err, ShouldBeNil)
		It(len(faker.Routers), ShouldEqual, 3)

		userModel := faker.Routers["users"].Model
		bookModel := faker.Routers["books"].Model
		avatarsModel := faker.Routers["avatars"].Model
		It(userModel.Len(), ShouldEqual, 3)
		It(bookModel.Len(), ShouldEqual, 3)
		It(avatarsModel.Len(), ShouldEqual, 1)

		li, _ := userModel.Get(float64(1))
		userModel.InsertRelatedData(&li)
		It(li.Len(), ShouldEqual, 6)
	})

	httpmock.ListenAndServe("localhost", faker)
	Describ("SetHandlers", t, func() {
		Describ("GET /users/:id", func() {
			Context("when pass a valid id\n", func() {
				response := httpmock.GET("/users/1", nil)
				It(response, shouldHasJsonResponse, usersFixture[0])
			})

			Context("when pass a invalid id\n", func() {
				response := httpmock.GET("/users/xxx", nil)
				It(response.Code, ShouldEqual, http.StatusBadRequest)
			})
		})

		Describ("GET /users", func() {
			response := httpmock.GET("/users", nil)
			It(response, shouldHasJsonResponse, usersFixture)
		})

		Describ("POST /users", func() {
			Context("when pass valid params", func() {
				response, _ := httpmock.POSTForm("/users", userParam)
				It(response.Code, ShouldEqual, http.StatusOK)

				respJSON := response.JSON()
				resMap, ok := respJSON.(map[string]interface{})
				It(ok, ShouldBeTrue)
				It(resMap["id"], ShouldEqual, float64(4))
				It(resMap["name"], ShouldEqual, userParam["name"])
				It(resMap["phone"], ShouldEqual, userParam["phone"])
				It(resMap["age"], ShouldEqual, (userParam["age"]))
				It(faker.Routers["users"].Model.Has(4), ShouldBeTrue)
			})

			Context("when pass invalid params", func() {
				response, _ := httpmock.POSTForm("/users", invalidUserParam)
				It(response.Code, ShouldEqual, http.StatusBadRequest)
			})
		})

		Describ("PATCH /users/:id", func() {
			Context("when pass valid params", func() {
				userEditedAttrPatch := map[string]interface{}{"name": "Vincent", "age": "22"}
				response, _ := httpmock.PATCH("/users/4", userEditedAttrPatch)
				It(response.Code, ShouldEqual, http.StatusOK)

				respJSON := response.JSON()
				resMap, ok := respJSON.(map[string]interface{})
				It(ok, ShouldBeTrue)
				It(resMap["id"], ShouldEqual, float64(4))
				It(resMap["name"], ShouldEqual, userEditedAttrPatch["name"])
				It(resMap["phone"], ShouldEqual, userParam["phone"])
				It(resMap["age"], ShouldEqual, userParam["age"])
			})

			Context("when pass invalid params", func() {
				response, _ := httpmock.PATCH("/users/4", invalidUserParam)
				It(response.Code, ShouldEqual, http.StatusBadRequest)
			})
		})

		Describ("PUT /users/:id", func() {
			userEditedAttrPut := map[string]interface{}{"name": "Vincent", "phone": "13213213217", "age": float64(23)}
			Context("when pass valid params", func() {
				response, _ := httpmock.PUT("/users/4", userEditedAttrPut)
				It(response.Code, ShouldEqual, http.StatusOK)

				respJSON := response.JSON()
				resMap, ok := respJSON.(map[string]interface{})
				It(ok, ShouldBeTrue)
				It(resMap["id"], ShouldEqual, float64(4))
				It(resMap["name"], ShouldEqual, userEditedAttrPut["name"])
				It(resMap["phone"], ShouldEqual, userEditedAttrPut["phone"])
				It(resMap["age"], ShouldEqual, userEditedAttrPut["age"])
			})

			Context("when pass invalid params", func() {
				response, _ := httpmock.PUT("/users/4", userEditedAttrPut)
				It(response.Code, ShouldEqual, http.StatusBadRequest)
			})
		})

		Describ("DELETE /users/:id", func() {
			response := httpmock.DELETE("/users/1", nil)
			It(response.Code, ShouldEqual, http.StatusOK)
			It(faker.Routers["users"].Model.Has(3), ShouldBeTrue)
			It(faker.Routers["books"].Model.Set.Len(), ShouldEqual, 1)
			It(faker.Routers["books"].Model.Has(float64(2)), ShouldBeTrue)
			It(faker.Routers["avatars"].Model.Len(), ShouldEqual, 0)
		})
	})

	Describ("MountToAndIntegrateHandler", t, func() {
		mux := http.NewServeMux()
		mux.HandleFunc("/greet", func(rw http.ResponseWriter, req *http.Request) {
			rw.WriteHeader(http.StatusOK)
			rw.Write([]byte("hello world"))
		})

		faker.MountTo("/fake_api")
		faker.IntegrateHandler(mux)

		response := httpmock.GET("/greet", nil)
		It(response.Body.String(), ShouldEqual, "hello world")

		response = httpmock.GET("/fake_api/users/2", nil)
		It(response, shouldHasJsonResponse, usersFixture[1])
	})
}
