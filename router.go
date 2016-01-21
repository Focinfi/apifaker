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
	// Name request parameter' name
	Name string
	// Desc description of this parameter
	Desc string
}

type Route struct {
	// Method request method only support "GET", "POST", "PUT", "DELETE"
	Method string

	// Path path of request's url
	Path string

	// Params array of this route
	Params []Param

	// Response a object of any thing can be json.Marshal
	Response interface{}
}

type Router struct {
	Resource string
	Routes   []Route
}

// newRouter create a new Router with reading a json file
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
