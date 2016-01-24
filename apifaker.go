package apifaker

import (
	"github.com/Focinfi/gset"
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

type ApiFaker struct {
	// Engine composing pointer of gin.Engine to let ApiFaker
	// has the ability of routing
	*gin.Engine

	// ApiDir the directory contains api json files
	ApiDir string

	// Routers a array of pointer of Router
	Routers []*Router

	// TrueMux external mux for the real api
	TrueMux http.Handler

	// Prefix the prefix of fake apis
	Prefix string
}

// NewWithApiDir create a new ApiFaker with the ApiDir,
// returns the pointer of ApiFaker and error will not be nil if
// there are some errors of manipulating ApiDir or json.Unmarshal
func NewWithApiDir(dir string) (*ApiFaker, error) {
	faker := &ApiFaker{Engine: gin.Default(), ApiDir: dir, Routers: []*Router{}}
	err := filepath.Walk(dir, func(path string, f os.FileInfo, err error) error {
		if f == nil {
			return err
		}
		if f.IsDir() {
			return nil
		}
		var router *Router
		if strings.HasSuffix(path, ".json") {
			if router, err = NewRouterWithPath(path); err == nil {
				faker.Routers = append(faker.Routers, router)
			}
		}
		return err
	})

	if err == nil {
		faker.setHandlers("")
	}
	return faker, err
}

// ServeHTTP serve the req and write response to rw.
// It will use gin.Engine when req.URL.Path hasing prefix of af.Prefix
// otherwise it will call af.TrueMux.ServeHTTP()
func (af *ApiFaker) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	path := req.URL.Path
	if af.Prefix == "" || strings.HasPrefix(path, af.Prefix+"/") {
		af.Engine.ServeHTTP(rw, req)
	} else {
		if af.TrueMux != nil {
			af.TrueMux.ServeHTTP(rw, req)
		}
	}
}

// MountTo assign path as af's Prefix and assign handler as af's TrueMux.
// At same time set hte handlers for af
func (af *ApiFaker) MountTo(path string, handler http.Handler) {
	af.Prefix = path
	af.Engine = gin.Default()
	af.setHandlers(path)
	af.TrueMux = handler
}

// setHandlers set all handlers into af.Engine.
// It will panic when there are has repeat combination of route.Method and route.Path
func (af *ApiFaker) setHandlers(prefix string) {
	for _, router := range af.Routers {
		for _, route := range router.Routes {
			method := route.Method
			resource := router.Resource
			path := prefix + route.Path
			switch method {
			case GET:
				af.GET(path, func(ctx *gin.Context) {
					idStr := ctx.Param("id")
					// GET /collection
					if idStr == "" {
						resourceSlice := NewItemsFromInterfaces(resource.Set.ToSlice())
						sort.Sort(resourceSlice)
						ctx.JSON(http.StatusOK, resourceSlice)
					} else if id, err := strconv.Atoi(idStr); err == nil {
						// GET /collection/:id
						if item, ok := resource.Get(id); ok {
							ctx.JSON(http.StatusOK, item)
						} else {
							ctx.JSON(http.StatusNotFound, nil)
						}
					} else {
						// GET /collection/xxx
						ctx.JSON(http.StatusBadRequest, nil)
					}
				})
			case POST:
				af.POST(path, func(ctx *gin.Context) {
					item, err := NewItemWithGinContext(ctx, resource)
					if err != nil {
						ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
					} else {
						if err = resource.Add(item); err != nil {
							ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
						} else {
							ctx.JSON(http.StatusOK, item)
						}
					}
				})
			case PUT:
				af.PUT(path, func(ctx *gin.Context) {
					idStr := ctx.Param("id")
					// check if id is int
					id, err := strconv.Atoi(idStr)
					if err != nil {
						ctx.JSON(http.StatusBadRequest, gin.H{"message": err})
						return
					}

					// update
					code, resp := resource.UpdateWithAllAttrsInGinContex(id, ctx)
					ctx.JSON(code, resp)
				})
			case PATCH:
				af.PATCH(path, func(ctx *gin.Context) {
					idStr := ctx.Param("id")
					// check if id is int
					id, err := strconv.Atoi(idStr)
					if err != nil {
						ctx.JSON(http.StatusBadRequest, gin.H{"message": err})
						return
					}

					// update
					code, resp := resource.UpdateWithAttrsInGinContext(id, ctx)
					ctx.JSON(code, resp)
				})
			case DELETE:
				af.DELETE(path, func(ctx *gin.Context) {
					idStr := ctx.Param("id")
					// check if id is int
					id, err := strconv.Atoi(idStr)
					if err != nil {
						ctx.JSON(http.StatusBadRequest, gin.H{"message": err})
						return
					}

					// delete
					resource.Remove(gset.T(id))
					ctx.JSON(http.StatusOK, nil)
				})
			}
		}
	}
}
