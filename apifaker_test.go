package apifaker

import (
	"github.com/Focinfi/gtester"
	"github.com/Focinfi/gtester/httpmock"
	// "net/http"
	"os"
	"testing"
)

var testDir = os.Getenv("GOPATH") + "/src/github.com/Focinfi/apifaker/api_static_test"

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

	response := httpmock.GET("/users/1", nil)
	gtester.AssertResponseEqual(t, response, usersFixture[0])

	response = httpmock.GET("/users", nil)
	gtester.AssertResponseEqual(t, response, usersFixture)

}

// func TestMountTo(t *testing.T) {
// 	faker, _ := NewWithApiDir(testDir)
// 	mux := http.NewServeMux()
// 	mux.HandleFunc("/greet", func(rw http.ResponseWriter, req *http.Request) {
// 		rw.WriteHeader(http.StatusOK)
// 		rw.Write([]byte("hello world"))
// 	})

// 	faker.MountTo("/fake_api", mux)
// 	httpmock.ListenAndServe("localhost", faker)

// 	response := httpmock.GET("/greet", nil)
// 	gtester.AssertEqual(t, response.Body.String(), "hello world")

// 	response = httpmock.GET("/fake_api/users/1", nil)
// 	gtester.AssertEqual(t, response.Code, http.StatusOK)
// 	expResp := map[string]string{"name": "Frank"}
// 	gtester.AssertResponseEqual(t, response, expResp)
// }
