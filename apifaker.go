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
	// Engine in charge of serving http requests
	*gin.Engine

	// ApiDir the directory contains api json files
	ApiDir string

	// Routers contains all routes use their name as the key
	Routers map[string]*Router

	// ExtMux the external mux for the real api
	ExtMux http.Handler

	// Prefix the prefix of fake apis
	Prefix string
}

// NewWithApiDir alloactes and returns a new ApiFaker with the given dir as its ApiDir,
// the error will not be nil if
//   1. dir is wrong
//   2. json file format is wrong
//   3. break rules described in README.md
func NewWithApiDir(dir string) (*ApiFaker, error) {
	faker := &ApiFaker{
		ApiDir:  dir,
		Routers: map[string]*Router{},
	}

	err := gtester.NewInspector().Check(func() error {
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
					return JsonFileErrorf("%s has been existed", router.Model.Name)
				} else {
					faker.Routers[router.Model.Name] = router
				}
			}
			return nil
		})
	}).Check(func() error {
		return faker.CheckUniqueness()
	}).Check(func() error {
		return faker.CheckRelationships()
	}).Then(func() {
		faker.setHandlers()
		faker.setSaveToFileTimer()
	})

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

// ServeHTTP implements the http.Handler.
// It will use Engine when req.URL.Path hasing prefix of Prefix or ExtMux is nil
// otherwise it will call ApiFaker.ExtMux.ServeHTTP()
func (af *ApiFaker) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	path := req.URL.Path
	if af.Prefix == "" || strings.HasPrefix(path, af.Prefix+"/") || af.ExtMux == nil {
		af.Engine.ServeHTTP(rw, req)
	} else {
		af.ExtMux.ServeHTTP(rw, req)
	}
}

// MountTo assign path as ApiFaker's Prefix and reset the handlers
func (af *ApiFaker) MountTo(path string) {
	af.Prefix = path
	af.setHandlers()
}

// IntegrateHandler set ApiFaker's ExtMux
func (af *ApiFaker) IntegrateHandler(handler http.Handler) {
	af.ExtMux = handler
}

// SaveToFile
func (af *ApiFaker) SaveToFile() {
	for _, router := range af.Routers {
		router.SaveToFile()
	}
}

// setSaveToFileTimer set a timer to call SaveToFile() once a day
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

// setHandlers set all handlers into ApiFaker.Engine.
func (af *ApiFaker) setHandlers() {
	// if panic, backfill data to json files
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
						newLi := li.InsertRelatedData(model)
						ctx.JSON(http.StatusOK, newLi.ToMap())
					} else {
						// GET /collection
						models := model.ToLineItems()
						sort.Sort(models)
						ctx.JSON(http.StatusOK, models.ToSlice())
					}
				})
			case POST:
				af.POST(path, func(ctx *gin.Context) {
					li, err := NewLineItemWithGinContext(ctx, model)
					if err == nil {
						err = model.Add(li)
					}

					if err != nil {
						ctx.JSON(http.StatusBadRequest, ResponseErrorMsg(err))
					} else {
						ctx.JSON(http.StatusOK, li.ToMap())
					}
				})
			case PUT:
				af.PUT(path, func(ctx *gin.Context) {
					// allocate a new item
					newLi, err := NewLineItemWithGinContext(ctx, model)

					if err != nil {
						ctx.JSON(http.StatusBadRequest, ResponseErrorMsg(err))
						return
					}

					// update
					id, _ := ctx.Get("idFloat64")
					if err := model.Update(id.(float64), &newLi); err != nil {
						ctx.JSON(http.StatusBadRequest, ResponseErrorMsg(err))
					} else {
						ctx.JSON(http.StatusOK, newLi.ToMap())
					}
				})
			case PATCH:
				af.PATCH(path, func(ctx *gin.Context) {
					// update with attrs, got error if attrs is not complete
					id, _ := ctx.Get("idFloat64")
					if li, err := model.UpdateWithAttrs(id.(float64), ctx); err != nil {
						ctx.JSON(http.StatusBadRequest, ResponseErrorMsg(err))
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

// NewGinEngineWithFaker allocate and returns a new gin.Engine pointer,
// added a new middleware which will check the type id param and the resource existence,
// if ok, set the float64 value of id named idFloat64, otherwise response 404 or 400.
func NewGinEngineWithFaker(faker *ApiFaker) *gin.Engine {
	engine := gin.Default()
	gin.SetMode(gin.ReleaseMode)
	// check id
	engine.Use(func(ctx *gin.Context) {
		// check if param "id" is int
		idStr := ctx.Param("id")
		if idStr == "" {
			return
		}

		id, err := strconv.ParseFloat(idStr, 64)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, ResponseErrorMsg(err))
		}

		path := strings.TrimSuffix(ctx.Request.URL.Path, "/")
		pathPieces := strings.Split(path, "/")
		resourceName := pathPieces[len(pathPieces)-2]

		if router, ok := faker.Routers[resourceName]; ok {
			if _, ok := router.Model.Get(id); ok {
				ctx.Set("idFloat64", id)
			} else {
				ctx.JSON(http.StatusNotFound, nil)
			}
		}
	})

	return engine
}
