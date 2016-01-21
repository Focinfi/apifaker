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
	TrueMux http.Handler
	Prefix  string
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
	return faker, err
}

func (af *ApiFaker) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	path := req.URL.Path
	if af.Prefix != "" && strings.HasPrefix(path, af.Prefix+"/") {
		af.Engine.ServeHTTP(rw, req)
	} else {
		af.TrueMux.ServeHTTP(rw, req)
	}
}

func (af *ApiFaker) MountTo(path string, handler http.Handler) {
	af.Prefix = path
	af.setHandlers(path)
	af.TrueMux = handler
}

func (af *ApiFaker) setHandlers(prefix string) {
	for _, router := range af.Routers {
		for _, route := range router.Routes {
			method := strings.ToUpper(route.Method)
			response := route.Response
			path := prefix + route.Path
			switch method {
			case "GET":
				af.GET(path, func(ctx *gin.Context) {
					ctx.JSON(http.StatusOK, gin.H{"data": response})
				})
			case "POST":
				af.POST(path, func(ctx *gin.Context) {
					ctx.JSON(http.StatusOK, gin.H{"data": response})
				})
			case "PUT":
				af.PUT(path, func(ctx *gin.Context) {
					ctx.JSON(http.StatusOK, gin.H{"data": response})
				})
			case "DELETE":
				af.DELETE(path, func(ctx *gin.Context) {
					ctx.JSON(http.StatusOK, gin.H{"data": response})
				})
			}
		}
	}
}
