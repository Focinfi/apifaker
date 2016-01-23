package apifaker

import (
	"github.com/Focinfi/gtester"
	"github.com/Focinfi/gtester/httpmock"
	// "net/http"
	"os"
	"testing"
)

var testDir = os.Getenv("GOPATH") + "/src/github.com/Focinfi/apifaker/api_static_test"

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
	expResp := map[string]interface{}{"id": 1, "name": "Frank", "phone": "13213213213", "age": 22}
	gtester.AssertResponseEqual(t, response, expResp)
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
