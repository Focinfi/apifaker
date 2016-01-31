package apifaker

import (
	"github.com/Focinfi/gset"
	. "github.com/Focinfi/gtester"
	"github.com/Focinfi/gtester/httpmock"
	"net/http"
	"os"
	"testing"
)

var testDir = os.Getenv("GOPATH") + "/src/github.com/Focinfi/apifaker/api_static_test"

var userParam = map[string]interface{}{
	"name": "Ameng", "phone": "13213213214", "age": float64(22),
}

var invalidUserParam = map[string]interface{}{
	"name": ",,,", "phone": "13013213214", "age": "22",
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

func TestNewWithApiDir(t *testing.T) {
	if faker, err := NewWithApiDir(testDir); err != nil {
		t.Error(err)
	} else {
		AssertEqual(t, len(faker.Routers), 2)
		AssertEqual(t, faker.Routers["users"].Model.Set.Len(), 3)
		AssertEqual(t, faker.Routers["books"].Model.Set.Len(), 3)
		userModel := faker.Routers["users"].Model
		li, _ := userModel.Get(float64(1))
		userModel.InsertRelatedData(&li)
		AssertEqual(t, li.Len(), 5)
	}
}

func TestSetHandlers(t *testing.T) {
	faker, _ := NewWithApiDir(testDir)
	httpmock.ListenAndServe("localhost", faker)

	// GET /users/:id
	response := httpmock.GET("/users/1", nil)
	AssertResponseEqual(t, response, usersFixture[0])

	// GET /users
	response = httpmock.GET("/users", nil)
	AssertResponseEqual(t, response, usersFixture)

	// GET /users/xxx
	response = httpmock.GET("/users/xxx", nil)
	AssertEqual(t, response.Code, http.StatusBadRequest)

	// POST /users
	// 	with valid params
	response, _ = httpmock.POSTForm("/users", userParam)
	AssertEqual(t, response.Code, http.StatusOK)

	respJSON := response.JSON()
	if resMap, ok := respJSON.(map[string]interface{}); ok {
		AssertEqual(t, resMap["id"], float64(4))
		AssertEqual(t, resMap["name"], userParam["name"])
		AssertEqual(t, resMap["phone"], userParam["phone"])
		AssertEqual(t, resMap["age"], (userParam["age"]))
	} else {
		t.Errorf("can not set Post handlers, response body is: %s", response.Body.String())
	}
	AssertEqual(t, faker.Routers["users"].Model.Set.Has(gset.T(float64(4))), true)

	// 	with invalid params
	response, _ = httpmock.POSTForm("/users", invalidUserParam)
	AssertEqual(t, response.Code, http.StatusBadRequest)

	// PATCH /users/:id
	// 	with valid params
	userEditedAttrPatch := map[string]interface{}{"name": "Vincent"}
	response, _ = httpmock.PATCH("/users/4", userEditedAttrPatch)
	AssertEqual(t, response.Code, http.StatusOK)

	respJSON = response.JSON()
	if resMap, ok := respJSON.(map[string]interface{}); ok {
		AssertEqual(t, resMap["id"], float64(4))
		AssertEqual(t, resMap["name"], userEditedAttrPatch["name"])
		AssertEqual(t, resMap["phone"], userParam["phone"])
		AssertEqual(t, resMap["age"], userParam["age"])
	} else {
		t.Errorf("can not set PATCH handlers, response body is: %s", response.Body.String())
	}

	// 	with invalid params
	response, _ = httpmock.PATCH("/users/4", invalidUserParam)
	AssertEqual(t, response.Code, http.StatusBadRequest)

	// PUT /users/:id
	// 	with valid params
	userEditedAttrPut := map[string]interface{}{"name": "Vincent", "phone": "13213213217", "age": float64(23)}
	response, _ = httpmock.PUT("/users/4", userEditedAttrPut)
	AssertEqual(t, response.Code, http.StatusOK)
	t.Logf("[PUT]%s", response.Body)

	respJSON = response.JSON()
	if resMap, ok := respJSON.(map[string]interface{}); ok {
		AssertEqual(t, resMap["id"], float64(4))
		AssertEqual(t, resMap["name"], userEditedAttrPut["name"])
		AssertEqual(t, resMap["phone"], userEditedAttrPut["phone"])
		AssertEqual(t, resMap["age"], userEditedAttrPut["age"])
	} else {
		t.Errorf("can not set PATCH handlers, response body is: %s", response.Body.String())
	}

	// with invalid params
	response, _ = httpmock.PUT("/users/4", userEditedAttrPut)
	AssertEqual(t, response.Code, http.StatusBadRequest)

	// DELETE /users/:id
	response = httpmock.DELETE("/users/1", nil)
	AssertEqual(t, response.Code, http.StatusOK)
	AssertEqual(t, faker.Routers["users"].Model.Has(3), true)
	AssertEqual(t, faker.Routers["books"].Model.Set.Len(), 1)
	AssertEqual(t, faker.Routers["books"].Model.Has(float64(2)), true)
}

func TestMountTo(t *testing.T) {
	faker, _ := NewWithApiDir(testDir)
	mux := http.NewServeMux()
	mux.HandleFunc("/greet", func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusOK)
		rw.Write([]byte("hello world"))
	})

	faker.MountTo("/fake_api", mux)
	httpmock.ListenAndServe("localhost", faker)

	response := httpmock.GET("/greet", nil)
	AssertEqual(t, response.Body.String(), "hello world")

	response = httpmock.GET("/fake_api/users/1", nil)
	AssertEqual(t, response.Code, http.StatusOK)
	AssertResponseEqual(t, response, usersFixture[0])
}
