package apifaker

import (
	"os"
	"testing"
)

func TestNewWithApiDir(t *testing.T) {
	pwd, _ := os.Getwd()
	if faker, err := NewWithApiDir(pwd + "/api_static_test"); err != nil {
		t.Error(err)
	} else {
		t.Logf("%#v", faker.Routers)
	}
}
