package apifaker

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

var tempRouter = Router{
	Resource: "books",
	Routes: []Route{
		{
			Method:   "GET",
			Path:     "/:id",
			Params:   []Param{},
			Response: `xxx`,
		},
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
			if len(r.Routes) == 0 {
				t.Errorf("can not create a new router, result is: %#v", r)
			}
		} else {
			t.Error("fail to create tempt file")
		}
	} else {
		t.Error("fail to marshal tempRouter")
	}
}
