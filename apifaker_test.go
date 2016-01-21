package apifaker

import (
	"github.com/Focinfi/gtester"
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
	gtester.ListenAndServe("localhost", faker)
	response := gtester.NewRecorder()

	gtester.Get("/book/1", response)
	gtester.AssertEqual(t, response.Code, http.StatusOK)
	expResp := map[string]map[string]string{"data": map[string]string{"title": "The Litle Prince"}}
	gtester.AssertJsonEqual(t, response, expResp)
}
