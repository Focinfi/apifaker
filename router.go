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
	Model  *Model
	Routes []Route

	filePath string `json:"-"`
}

func (r *Router) setRestRoutes() {
	r.Routes = []Route{
		// GET /collection
		{GET, fmt.Sprintf("/%s", r.Model.Name)},

		// GET /collection/:id
		{GET, fmt.Sprintf("/%s/:id", r.Model.Name)},

		// POST /collection
		{POST, fmt.Sprintf("/%s", r.Model.Name)},

		// PUT /collection
		{PUT, fmt.Sprintf("/%s/:id", r.Model.Name)},

		// PATCH /collection
		{PATCH, fmt.Sprintf("/%s/:id", r.Model.Name)},

		// DELETE /collection
		{DELETE, fmt.Sprintf("/%s/:id", r.Model.Name)},
	}

}

func (r *Router) SaveToFile() error {
	return r.Model.SaveToFile(r.filePath)
}

func NewRouterWithPath(path string) (*Router, error) {
	model, err := NewModelWithPath(path)
	if err != nil {
		return nil, err
	}

	router := &Router{Model: model, filePath: path}
	router.setRestRoutes()
	return router, err
}
