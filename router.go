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

type Route struct {
	// Method request method only support GET, POST, PUT, PATCH, DELETE
	Method RestMethod

	// Path
	Path string
}

type Router struct {
	Model  *Model
	Routes []Route

	apiFaker *ApiFaker
	filePath string
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

// SaveToFile
func (r *Router) SaveToFile() error {
	return r.Model.SaveToFile(r.filePath)
}

// NewRouterWithPath allocates and returns a new Router with the givn file path
func NewRouterWithPath(path string, apiFaker *ApiFaker) (*Router, error) {
	router := &Router{apiFaker: apiFaker, filePath: path}
	model, err := NewModelWithPath(path, router)
	if err != nil {
		return nil, err
	}

	router.Model = model
	router.setRestRoutes()
	return router, err
}
