package apifaker

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type ApiFaker struct {
	*gin.Engine
	ApiDir  string
	Routers []*Router
}

func NewWithApiDir(dir string) (*ApiFaker, error) {
	faker := ApiFaker{Engine: gin.Default(), ApiDir: dir, Routers: []*Router{}}
	err := filepath.Walk(dir, func(path string, f os.FileInfo, err error) error {
		if f == nil {
			return err
		}
		if f.IsDir() {
			return nil
		}
		var router *Router
		if strings.HasSuffix(path, ".json") {
			if router, err = newRouter(path); err == nil {
				faker.Routers = append(faker.Routers, router)
			}
		}
		return err
	})

	if err == nil {
		faker.setHandlers()
	}
	return &faker, err
}

func (af *ApiFaker) setHandlers() {
	for _, router := range af.Routers {
		for _, route := range router.Routes {
			method := strings.ToUpper(route.Method)
			switch method {
			case "GET":
				fmt.Printf("[Route] %#v\n", route)
				af.GET(route.Path, func(ctx *gin.Context) {
					ctx.JSON(http.StatusOK, gin.H{"data": route.Response})
				})
			case "POST":
				af.POST(route.Path, func(ctx *gin.Context) {
					ctx.JSON(http.StatusOK, gin.H{"data": route.Response})
				})
			case "PUT":
				af.PUT(route.Path, func(ctx *gin.Context) {
					ctx.JSON(http.StatusOK, gin.H{"data": route.Response})
				})
			case "DELETE":
				af.DELETE(route.Path, func(ctx *gin.Context) {
					ctx.JSON(http.StatusOK, gin.H{"data": route.Response})
				})
			}
		}
	}
}

func (af *ApiFaker) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	af.Engine.ServeHTTP(rw, req)
}
