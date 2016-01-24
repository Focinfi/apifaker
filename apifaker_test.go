package apifaker

import (
	"github.com/Focinfi/gset"
	"github.com/Focinfi/gtester"
	"github.com/Focinfi/gtester/httpmock"
	"net/http"
	"os"
	"testing"
)

var testDir = os.Getenv("GOPATH") + "/src/github.com/Focinfi/apifaker/api_static_test"

var userParam = map[string]interface{}{
	"name": "Ameng", "phone": "13213213214", "age": "22",
}

var usersFixture = []map[string]interface{}{
	{"id": 1, "name": "Frank", "phone": "13213213213", "age": 22},
	{"id": 2, "name": "Antony", "phone": "13213213211", "age": 22},
	{"id": 3, "name": "Foci", "phone": "13213213212", "age": 22},
}

func TestNewWithApiDir(t *testing.T) {
	if faker, err := NewWithApiDir(testDir); err != nil {
		t.Error(err)
	} else {
		t.Logf("%#v", faker.Routers[0])
	}
}

func TestSetHandlers(t *testing.T) {
	faker, _ := NewWithApiDir(testDir)
	httpmock.ListenAndServe("localhost", faker)

	// GET /users/:id
	response := httpmock.GET("/users/1", nil)
	gtester.AssertResponseEqual(t, response, usersFixture[0])

	// GET /users
	response = httpmock.GET("/users", nil)
	gtester.AssertResponseEqual(t, response, usersFixture)

	// GET /users/xxx
	response = httpmock.GET("/users/xxx", nil)
	gtester.AssertEqual(t, response.Code, http.StatusBadRequest)

	// POST /users
	response, _ = httpmock.POSTForm("/users", userParam)
	gtester.AssertEqual(t, response.Code, http.StatusOK)

	respJSON := response.JSON()
	if resMap, ok := respJSON.(map[string]interface{}); ok {
		gtester.AssertEqual(t, resMap["id"], float64(4))
		gtester.AssertEqual(t, resMap["name"], userParam["name"])
		gtester.AssertEqual(t, resMap["phone"], userParam["phone"])
		gtester.AssertEqual(t, resMap["age"], (userParam["age"]))
	} else {
		t.Errorf("can not set Post handlers, response body is: %s", response.Body.String())
	}

	gtester.AssertEqual(t, faker.Routers[0].Resource.Set.Has(gset.T(4)), true)

	// PUT /users/:id

	// PATCH /users/:id
	userEditedAttr := map[string]interface{}{"name": "Vincent"}
	response, _ = httpmock.PATCH("/users/4", userEditedAttr)
	gtester.AssertEqual(t, response.Code, http.StatusOK)

	respJSON = response.JSON()
	if resMap, ok := respJSON.(map[string]interface{}); ok {
		gtester.AssertEqual(t, resMap["id"], float64(4))
		gtester.AssertEqual(t, resMap["name"], userEditedAttr["name"])
		gtester.AssertEqual(t, resMap["phone"], userParam["phone"])
		gtester.AssertEqual(t, resMap["age"], (userParam["age"]))
	} else {
		t.Errorf("can not set PATCH handlers, response body is: %s", response.Body.String())
	}
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
	gtester.AssertEqual(t, response.Body.String(), "hello world")

	response = httpmock.GET("/fake_api/users/1", nil)
	gtester.AssertEqual(t, response.Code, http.StatusOK)
	gtester.AssertResponseEqual(t, response, usersFixture[0])
}
