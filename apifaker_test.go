package apifaker

import (
	"github.com/Focinfi/gtester"
	"github.com/Focinfi/gtester/httpmock"
	"net/http"
	"os"
	"testing"
)

var pwd, _ = os.Getwd()

func TestNewWithApiDir(t *testing.T) {
	if faker, err := NewWithApiDir(pwd + "/api_static_test"); err != nil {
		t.Error(err)
	} else {
		t.Logf("%#v", faker.Routers)
	}
}

func TestSetHandlers(t *testing.T) {
	faker, _ := NewWithApiDir(pwd + "/api_static_test")
	mux := http.NewServeMux()
	mux.HandleFunc("/greet", func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusOK)
		rw.Write([]byte("hello world"))
	})

	faker.MountTo("/fake_api", mux)
	httpmock.ListenAndServe("localhost", faker)

	response := httpmock.GET("/greet", nil)
	gtester.AssertEqual(t, response.Body.String(), "hello world")

	response = httpmock.GET("/fake_api/user/1", nil)
	gtester.AssertEqual(t, response.Code, http.StatusOK)
	expResp := map[string]map[string]string{"data": map[string]string{"name": "Frank"}}
	gtester.AssertResponseEqual(t, response, expResp)
}
