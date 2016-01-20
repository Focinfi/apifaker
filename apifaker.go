package apifaker

import (
	// "fmt"
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

	return &faker, err
}

func (af *ApiFaker) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	af.ServeHTTP(rw, req)
}
