package app

import (
	"bytes"
	"log"
	"net/http"
	"runtime"
	"runtime/pprof"
	"sync/atomic"
	"time"

	"github.com/coocood/freecache"
	"github.com/valyala/fasthttp"

	"github.com/ei-grad/hlcup/db"
	"github.com/ei-grad/hlcup/entities"
)

// Application implements application logic
type Application struct {
	db            *db.DB
	cache         *freecache.Cache
	countRequests int32
	heat          func(entities.Entity, uint32)
}

// NewApplication creates new Application
func NewApplication() *Application {
	var app Application
	app.db = db.New()
	app.cache = freecache.NewCache(512 * 1024 * 1024)
	return &app
}

func (app *Application) RpsWatcher() {
	for {
		time.Sleep(1 * time.Second)
		count := atomic.LoadInt32(&app.countRequests)
		if count > 0 {
			log.Printf("RPS: %6d | CACHE HIT RATE: %6d / %6d | CACHED ENTRIES %d", count,
				app.cache.HitCount(), app.cache.LookupCount(), app.cache.EntryCount())
			app.cache.ResetStatistics()
			atomic.SwapInt32(&app.countRequests, 0)
		}
	}
}

func handlePprof(ctx *fasthttp.RequestCtx, entity []byte) int {
	if bytes.Equal(entity, []byte("pprof")) {
		if err := pprof.StartCPUProfile(ctx); err != nil {
			log.Print("could not start CPU profile: ", err)
			return http.StatusInternalServerError
		}
		time.Sleep(60 * time.Second)
		pprof.StopCPUProfile()
		return http.StatusOK
	}
	if bytes.Equal(entity, []byte("pprof_mem")) {
		runtime.GC() // get up-to-date statistics
		if err := pprof.WriteHeapProfile(ctx); err != nil {
			log.Fatal("could not write memory profile: ", err)
		}
		return http.StatusOK
	}
	return http.StatusNotFound
}

// RequestHandler contains routing implementation
func (app *Application) RequestHandler(ctx *fasthttp.RequestCtx) {

	atomic.AddInt32(&app.countRequests, 1)

	ctx.SetContentType("application/json; charset=utf8")

	var (
		id     uint32
		status int
		err    error
	)

	path := ctx.Request.Header.RequestURI()

	switch string(ctx.Method()) {

	case "GET":
		var entityEnd = 1
		for ; entityEnd < len(path); entityEnd++ {
			if path[entityEnd] == '/' {
				break
			}
		}
		entity := path[1:entityEnd]
		if entityEnd < len(path) {
			var idEnd = entityEnd + 1
			for ; idEnd < len(path); idEnd++ {
				if path[idEnd] == '/' {
					break
				}
			}
			idBytes := path[entityEnd+1 : idEnd]
			if idEnd == len(path) {
				id, err = parseUint32(idBytes)
				switch {
				case err == nil:
					// /<entity>/<id:int>
					if v, err := app.cache.Get(path); err == nil {
						// response from cache
						ctx.Write(v)
						return
					}
					status = app.GetEntity(ctx, entities.GetEntityByRoute(entity), id)
					if status == http.StatusOK {
						app.cache.Set(path, ctx.Response.Body(), 0)
					}
				case bytes.Equal(idBytes, []byte("new")):
					// /<entity>/new is POST-only, say 405 for convenience
					status = http.StatusMethodNotAllowed
				}
			} else {
				tailEnd := idEnd + 1
				for ; tailEnd < len(path); tailEnd++ {
					if path[tailEnd] == '/' || path[tailEnd] == '?' {
						break
					}
				}
				tail := path[idEnd+1 : tailEnd]
				if tailEnd == len(path) || path[tailEnd] == '?' {
					id, err = parseUint32(idBytes)
					if err == nil {
						e := entities.GetEntityByRoute(entity)
						switch {
						case e == entities.User && bytes.Equal(tail, bytesVisits):
							// /user/<id>/visits
							status = app.GetUserVisits(ctx, id, ctx.QueryArgs())
						case e == entities.Location && bytes.Equal(tail, bytesAvg):
							// /locations/<id>/avg
							status = app.GetLocationAvg(ctx, id, ctx.QueryArgs())
						case e == entities.Location && bytes.Equal(tail, bytesMarks):
							// /locations/<id>/marks - utility handler for debug
							status = app.GetLocationMarks(ctx, id)
						}
					}
				}
			}
		} else {
			// /pprof
			status = handlePprof(ctx, entity)
			break
		}
		if status == 0 {
			status = http.StatusNotFound
		}
	case "POST":

		// To fix the "Empty response" error in yandex-tank logs we have to send
		// "Connection: close" for POST requests.
		ctx.SetConnectionClose()

		// Also, check system expects a {} in the response body
		ctx.Write([]byte("{}"))

		var entityEnd = 1
		for ; entityEnd < len(path); entityEnd++ {
			if path[entityEnd] == '/' {
				break
			}
		}
		entity := path[1:entityEnd]
		if entityEnd < len(path) {
			var idEnd = entityEnd + 1
			for ; idEnd < len(path); idEnd++ {
				if path[idEnd] == '/' || path[idEnd] == '?' {
					break
				}
			}
			idBytes := path[entityEnd+1 : idEnd]
			if idEnd == len(path) || path[idEnd] == '?' {
				id, err = parseUint32(idBytes)
				switch {
				case err == nil:
					// /<entity>/<id:int>
					status = app.PostEntity(entities.GetEntityByRoute(entity), id, ctx.PostBody())
				case bytes.Equal(idBytes, []byte("new")):
					// /<entity>/new
					status = app.PostEntityNew(entities.GetEntityByRoute(entity), ctx.PostBody())
				}
			}
		}

	default:
		// XXX: rewrite with typed handlers to fix 405 errors on all urls?
		status = http.StatusMethodNotAllowed
	}

	ctx.SetStatusCode(status)

}
