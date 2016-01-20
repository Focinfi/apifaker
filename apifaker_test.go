package apifaker

import (
	"github.com/Focinfi/gtester"
	"net/http"
	"os"
	"strings"
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
	t.Log(faker.Engine)
	gtester.ListenAndServe("localhost", faker)
	response := gtester.NewRecorder()

	gtester.Get("/book/1", response)
	gtester.AssertEqual(t, response.Code, http.StatusOK)

	if !strings.EqualFold(response.Body.String(), `{"data":{"title":"The Litle Prince"}}`) {
		t.Errorf("can not set Handler, response body is:%s", response.Body.String())
	} else {
		t.Log("[Response]", response.Body.String())
	}
}
