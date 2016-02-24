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
var It = Convey
var Expect = So

var testDir = os.Getenv("GOPATH") + "/src/github.com/Focinfi/apifaker/api_static_test"

var userParam = map[string]interface{}{
	"id": float64(4), "name": "Ameng", "phone": "13213213214", "age": float64(22),
}

var incompleteParam = map[string]interface{}{"name": "N"}

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
	userModel := faker.Routers["users"].Model
	bookModel := faker.Routers["books"].Model
	avatarsModel := faker.Routers["avatars"].Model
	Describ("NewWithDir", t, func() {
		It("allocates and returns a new ApiFaker", func() {
			Expect(err, ShouldBeNil)
			Expect(len(faker.Routers), ShouldEqual, 3)
			Expect(userModel.Len(), ShouldEqual, 3)
			Expect(bookModel.Len(), ShouldEqual, 3)
			Expect(avatarsModel.Len(), ShouldEqual, 1)
		})
	})

	httpmock.ListenAndServe("localhost", faker)
	Describ("SetHandlers", t, func() {
		Describ("GET /users/:id", func() {
			Context("when pass a valid id", func() {
				response := httpmock.GET("/users/1", nil)
				It("returns the first user", func() {
					Expect(response, shouldHasJsonResponse, usersFixture[0])
				})
			})

			Context("when pass a invalid id", func() {
				response := httpmock.GET("/users/xxx", nil)
				It("returns 404", func() {
					Expect(response.Code, ShouldEqual, http.StatusBadRequest)
				})

				response = httpmock.GET("/users/100", nil)
				It("returns 400", func() {
					Expect(response.Code, ShouldEqual, http.StatusNotFound)
				})
			})
		})

		Describ("GET /users", func() {
			response := httpmock.GET("/users", nil)
			It("returns 200", func() {
				Expect(response, shouldHasJsonResponse, usersFixture)
			})
		})

		Describ("POST /users", func() {
			Context("when pass valid params", func() {
				response, _ := httpmock.POSTForm("/users", userParam)
				It("returns 200 and the created user", func() {
					Expect(response.Code, ShouldEqual, http.StatusOK)
					Expect(response, shouldHasJsonResponse, userParam)
					Expect(faker.Routers["users"].Model.Has(float64(4)), ShouldBeTrue)
				})
			})

			Context("when pass invalid params", func() {
				response, _ := httpmock.POSTForm("/users", invalidUserParam)
				It("returns 404", func() {
					Expect(response.Code, ShouldEqual, http.StatusBadRequest)
				})

				response, _ = httpmock.POSTForm("/users", incompleteParam)
				It("returns 404, params is incomplete", func() {
					Expect(response.Code, ShouldEqual, http.StatusBadRequest)
				})
			})
		})

		Describ("PATCH /users/:id", func() {
			Context("when pass valid params", func() {
				userEditedAttrPatch := map[string]interface{}{"name": "Vincent", "age": "22"}
				response, _ := httpmock.PATCH("/users/4", userEditedAttrPatch)
				respJSON := response.JSON()
				resMap, ok := respJSON.(map[string]interface{})
				It("returns 200 and the edited user", func() {
					Expect(response.Code, ShouldEqual, http.StatusOK)
					Expect(ok, ShouldBeTrue)
					Expect(resMap["id"], ShouldEqual, float64(4))
					Expect(resMap["name"], ShouldEqual, userEditedAttrPatch["name"])
					Expect(resMap["phone"], ShouldEqual, userParam["phone"])
					Expect(resMap["age"], ShouldEqual, userParam["age"])
				})
			})

			Context("when pass invalid params", func() {
				response, _ := httpmock.PATCH("/users/4", invalidUserParam)
				It("returns 404", func() {
					Expect(response.Code, ShouldEqual, http.StatusBadRequest)
				})
			})
		})

		Describ("PUT /users/:id", func() {
			userEditedAttrPut := map[string]interface{}{"id": float64(4), "name": "Vincent", "phone": "13213213217", "age": float64(23)}
			Context("when pass valid params", func() {
				response, _ := httpmock.PUT("/users/4", userEditedAttrPut)
				It("returns 200 and and the edited user", func() {
					Expect(response.Code, ShouldEqual, http.StatusOK)
					Expect(response, shouldHasJsonResponse, userEditedAttrPut)
				})
			})

			Context("when pass invalid params", func() {
				response, _ := httpmock.PUT("/users/4", userEditedAttrPut)
				It("returns 404", func() {
					Expect(response.Code, ShouldEqual, http.StatusBadRequest)
				})

				response, _ = httpmock.PUT("/users/4", incompleteParam)
				It("returns 404, params is incomplete", func() {
					Expect(response.Code, ShouldEqual, http.StatusBadRequest)
				})
			})
		})

		Describ("DELETE /users/:id", func() {
			response := httpmock.DELETE("/users/1", nil)
			It("returns 200 and delete all related resources", func() {
				Expect(response.Code, ShouldEqual, http.StatusOK)
				Expect(faker.Routers["users"].Model.Len(), ShouldEqual, 3)
				Expect(faker.Routers["books"].Model.Set.Len(), ShouldEqual, 1)
				Expect(faker.Routers["books"].Model.Has(float64(2)), ShouldBeTrue)
				Expect(faker.Routers["avatars"].Model.Len(), ShouldEqual, 0)
			})
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

		It("has a /greet route", func() {
			response := httpmock.GET("/greet", nil)
			Expect(response.Body.String(), ShouldEqual, "hello world")
		})
		It("adds a /apifaker prefix for internal routes", func() {
			response := httpmock.GET("/fake_api/users/2", nil)
			Expect(response, shouldHasJsonResponse, usersFixture[1])
		})
	})
}
