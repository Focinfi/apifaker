package apifaker

import (
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
			if router, err = newRouter(path); err == nil {
				faker.Routers = append(faker.Routers, router)
			}
		}
		return err
	})

	if err == nil {
		faker.setHandlers()
	}
	return faker, err
}

func (af *ApiFaker) setHandlers() {
	for _, router := range af.Routers {
		for _, route := range router.Routes {
			method := strings.ToUpper(route.Method)
			response := route.Response
			switch method {
			case "GET":
				af.GET(route.Path, func(ctx *gin.Context) {
					ctx.JSON(http.StatusOK, gin.H{"data": response})
				})
			case "POST":
				af.POST(route.Path, func(ctx *gin.Context) {
					ctx.JSON(http.StatusOK, gin.H{"data": response})
				})
			case "PUT":
				af.PUT(route.Path, func(ctx *gin.Context) {
					ctx.JSON(http.StatusOK, gin.H{"data": response})
				})
			case "DELETE":
				af.DELETE(route.Path, func(ctx *gin.Context) {
					ctx.JSON(http.StatusOK, gin.H{"data": response})
				})
			}
		}
	}
}

func (af *ApiFaker) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	af.Engine.ServeHTTP(rw, req)
}
