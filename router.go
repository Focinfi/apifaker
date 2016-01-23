package apifaker

import (
	"fmt"
)

type RestMethod int

const (
	GET RestMethod = 1 << (10 * iota)
	POST
	PUT
	PATCH
	DELETE
)

func (r RestMethod) String() string {
	switch r {
	case GET:
		return "GET"
	case POST:
		return "POST"
	case PUT:
		return "PUT"
	case PATCH:
		return "PATCH"
	case DELETE:
		return "DELETE"
	default:
		return ""
	}
	return ""
}

type Route struct {
	// Method request method only support GET, POST, PUT, PATCH, DELETE
	Method RestMethod

	// Path
	Path string
}

type Router struct {
	*Resource
	Routes []Route

	filePath string `json:"-"`
}

func (r *Router) setRestRoutes() {
	r.Routes = []Route{
		// GET /collection
		{GET, fmt.Sprintf("/%s", r.Name)},

		// GET /collection/:id
		{GET, fmt.Sprintf("/%s/:id", r.Name)},

		// POST /collection
		{POST, fmt.Sprintf("/%s", r.Name)},

		// PUT /collection
		{PUT, fmt.Sprintf("/%s/:id", r.Name)},

		// PATCH /collection
		{PATCH, fmt.Sprintf("/%s/:id", r.Name)},

		// DELETE /collection
		{DELETE, fmt.Sprintf("/%s/:id", r.Name)},
	}

}

func NewRouterWithPath(path string) (*Router, error) {
	resource, err := NewResourceWithPath(path)
	if err != nil {
		return nil, err
	}

	router := &Router{Resource: resource, filePath: path}
	router.setRestRoutes()
	return router, err
}
