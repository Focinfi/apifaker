package apifaker

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

var tempRouter = Router{
	Path: "/books",
	Get: Route{
		Params: []Param{
			Param{"id", "book's id"},
		},
		Response: `xxx`,
	},
}

func TestNewRouter(t *testing.T) {
	if bytes, err := json.Marshal(tempRouter); err == nil {
		if file, err := ioutil.TempFile("./", "router"); err == nil {
			defer file.Close()
			defer os.Remove(file.Name())
			t.Log(string(bytes))
			file.Write(bytes)
			path, err := filepath.Abs(file.Name())
			r, err := newRouter(path)
			if err != nil {
				t.Error(err)
			}
			if r.Get.Params[0].Name != "id" || r.Get.Params[0].Desc != "book's id" || r.Get.Response != "xxx" {
				t.Errorf("can not create a new router, result is: %#v", r)
			}
		} else {
			t.Error("fail to create tempt file")
		}
	} else {
		t.Error("fail to marshal tempRouter")
	}
}
