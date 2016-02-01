package apifaker

import (
	"github.com/Focinfi/gtester"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

type ApiFaker struct {
	// Engine composing pointer of gin.Engine to let ApiFaker
	// has the ability of routing
	*gin.Engine

	// ApiDir the directory contains api json files
	ApiDir string

	// Routers a array of pointer for Router
	Routers map[string]*Router

	// ExtMux external mux for the real api
	ExtMux http.Handler

	// Prefix the prefix of fake apis
	Prefix string
}

// NewWithApiDir create a new ApiFaker with the ApiDir,
// returns the pointer of ApiFaker and error will not be nil if
// there are some errors of manipulating ApiDir or json.Unmarshal
func NewWithApiDir(dir string) (*ApiFaker, error) {
	faker := &ApiFaker{
		ApiDir:  dir,
		Routers: map[string]*Router{},
	}

	err := gtester.NewCheckQueue().Add(func() error {
		// traverse dir
		return filepath.Walk(dir, func(path string, f os.FileInfo, err error) error {
			if f == nil {
				return err
			}
			if f.IsDir() || !strings.HasSuffix(path, ".json") {
				return nil
			}

			if router, err := NewRouterWithPath(path, faker); err != nil {
				return err
			} else {
				if _, ok := faker.Routers[router.Model.Name]; ok {
					panic(router.Model.Name + " has been existed")
				} else {
					faker.Routers[router.Model.Name] = router
				}
			}
			return nil
		})
	}).Add(func() error {
		return faker.CheckUniqueness()
	}).Add(func() error {
		return faker.CheckRelationships()
	}).Run()

	if err == nil {
		faker.setHandlers()
		faker.setSaveToFileTimer()
	}

	return faker, err
}

// CheckUniqueness
func (af *ApiFaker) CheckUniqueness() error {
	for _, router := range af.Routers {
		if err := router.Model.CheckUniqueness(); err != nil {
			return err
		}
	}
	return nil
}

// CheckRelationships
func (af *ApiFaker) CheckRelationships() error {
	for _, router := range af.Routers {
		if err := router.Model.CheckRelationships(); err != nil {
			return err
		}
	}
	return nil
}

// ServeHTTP serve the req and write response to rw.
// It will use gin.Engine when req.URL.Path hasing prefix of af.Prefix
// otherwise it will call af.ExtMux.ServeHTTP()
func (af *ApiFaker) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	path := req.URL.Path
	if af.Prefix == "" || strings.HasPrefix(path, af.Prefix+"/") || af.ExtMux == nil {
		af.Engine.ServeHTTP(rw, req)
	} else {
		af.ExtMux.ServeHTTP(rw, req)
	}
}

// MountTo assign path as af's Prefix and assign handler as af's ExtMux.
// At same time set hte handlers for af
func (af *ApiFaker) MountTo(path string, handler http.Handler) {
	af.Prefix = path
	af.setHandlers()
	af.ExtMux = handler
}

// SaveToFile
func (af *ApiFaker) SaveToFile() {
	for _, router := range af.Routers {
		router.SaveToFile()
	}
}

// setSaveToFileTimer set a periodic timer to call SaveToFile()
func (af *ApiFaker) setSaveToFileTimer() {
	go func() {
		for {
			now := time.Now()
			next := now.Add(time.Hour * 24)
			nextSaveDate := time.Date(next.Year(),
				next.Month(),
				next.Day(),
				0, 0, 0, 0,
				next.Location())
			timer := time.NewTimer(nextSaveDate.Sub(now))
			<-timer.C
			af.SaveToFile()
		}
	}()
}

// setHandlers set all handlers into af.Engine.
// It will panic when there are has repeat combination of route.Method and route.Path
func (af *ApiFaker) setHandlers() {
	// if any panic exsits, Save data to json
	defer func() {
		if err := recover(); err != nil {
			log.Fatal(err)
			af.SaveToFile()
		}
	}()

	// reset Engine
	af.Engine = NewGinEngineWithFaker(af)

	for _, router := range af.Routers {
		for _, route := range router.Routes {
			model := router.Model
			method := route.Method
			path := af.Prefix + route.Path
			switch method {
			case GET:
				af.GET(path, func(ctx *gin.Context) {
					if id, ok := ctx.Get("idFloat64"); ok {
						// GET /collection/:id
						li, _ := model.Get(id.(float64))
						model.InsertRelatedData(&li)
						ctx.JSON(http.StatusOK, li.ToMap())
					} else {
						// GET /collection
						models := model.ToLineItems()
						sort.Sort(models)
						ctx.JSON(http.StatusOK, models.ToSlice())
					}
				})
			case POST:
				af.POST(path, func(ctx *gin.Context) {
					// create a new lineitems with ctx.PostForm()
					li, err := NewLineItemWithGinContext(ctx, model)
					if err != nil {
						ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
					} else {
						// add this lineitem to model
						if err = model.Add(li); err != nil {
							ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
						} else {
							ctx.JSON(http.StatusOK, li.ToMap())
						}
					}
				})
			case PUT:
				af.PUT(path, func(ctx *gin.Context) {
					// allocate a new item
					newLi, err := NewLineItemWithGinContext(ctx, model)

					if err != nil {
						ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
						return
					}

					// update
					id, _ := ctx.Get("idFloat64")
					if err := model.Update(id.(float64), &newLi); err != nil {
						ctx.JSON(http.StatusBadRequest, map[string]string{"message": err.Error()})
					} else {
						ctx.JSON(http.StatusOK, newLi.ToMap())
					}
				})
			case PATCH:
				af.PATCH(path, func(ctx *gin.Context) {
					// update with attrs, got error if attrs is not complete
					id, _ := ctx.Get("idFloat64")
					if li, err := model.UpdateWithAttrs(id.(float64), ctx); err != nil {
						ctx.JSON(http.StatusBadRequest, map[string]string{"message": err.Error()})
					} else {
						ctx.JSON(http.StatusOK, li.ToMap())
					}
				})
			case DELETE:
				af.DELETE(path, func(ctx *gin.Context) {
					// delete
					id, _ := ctx.Get("idFloat64")
					model.Delete(id.(float64))
					ctx.JSON(http.StatusOK, nil)
				})
			}
		}
	}
}

func NewGinEngineWithFaker(faker *ApiFaker) *gin.Engine {
	engine := gin.Default()
	// check id
	engine.Use(func(ctx *gin.Context) {
		// check if param "id" is int
		idStr := ctx.Param("id")
		if idStr == "" {
			ctx.Next()
			return
		}

		id, err := strconv.ParseFloat(idStr, 64)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
			return
		}

		path := strings.TrimSuffix(ctx.Request.URL.Path, "/")
		pathPieces := strings.Split(path, "/")
		resourceName := pathPieces[len(pathPieces)-2]

		// check if element does exsit
		if router, ok := faker.Routers[resourceName]; ok {
			if _, ok := router.Model.Get(id); ok {
				ctx.Set("idFloat64", id)
				ctx.Next()
			} else {
				ctx.JSON(http.StatusNotFound, nil)
				return
			}
		} else {
			ctx.JSON(http.StatusNotFound, nil)
			return
		}
	})

	return engine
}
