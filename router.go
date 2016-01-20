package apifaker

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

type RestMethod int

const (
	GET RestMethod = 1 << (10 * iota)
	POST
	PUT
	DELETE
)

type Param struct {
	Name string
	Desc string
}

type Route struct {
	Method   string
	Path     string
	Params   []Param
	Response interface{}
}

type Router struct {
	Resource string
	Routes   []Route
}

func newRouter(apiPath string) (r *Router, err error) {
	file, err := os.Open(apiPath)
	if err != nil {
		return
	}
	defer file.Close()

	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		return
	}
	r = &Router{}
	if err = json.Unmarshal(bytes, r); err != nil {
		err = fmt.Errorf("[apifaker] json format error: %s, file: %s", err.Error(), apiPath)
	}
	return
}
