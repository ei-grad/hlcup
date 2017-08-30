package app

import (
	"bytes"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/valyala/fasthttp"

	"github.com/ei-grad/hlcup/db"
	"github.com/ei-grad/hlcup/entities"
)

// Application implements application logic
type Application struct {
	db            *db.DB
	now           time.Time
	countRequests int32
	heat          func(entities.Entity, uint32)
}

// NewApplication creates new Application
func NewApplication() *Application {
	var app Application
	app.db = db.New()
	return &app
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
			if path[entityEnd] == '/' || path[entityEnd] == '?' {
				break
			}
		}
		entity := path[1:entityEnd]
		if entityEnd < len(path) && path[entityEnd] != '?' {
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
					status = app.GetEntity(ctx, entities.GetEntityByRoute(entity), id)
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
						}
					}
				}
			}
		} else {
			// /pprof
			status = GetPprof(ctx, entity)
			break
		}
	case "POST":

		// To fix the "Empty response" error in yandex-tank logs we have to send
		// "Connection: close" for POST requests.
		// Fixed in test system, see #52
		//ctx.SetConnectionClose()

		// Also, check system expects a {} in the response body
		ctx.Write([]byte("{}"))

		body := ctx.PostBody()

		// XXX: some email's could contain the null string... but hopefully - not :-)
		if bytes.Contains(body, []byte(": null")) {
			//ctx.Logger().Printf("found null value: %s", body)
			status = http.StatusBadRequest
			break
		}

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
					status = app.PostEntity(entities.GetEntityByRoute(entity), id, body)
				case bytes.Equal(idBytes, []byte("new")):
					// /<entity>/new
					status = app.PostEntityNew(entities.GetEntityByRoute(entity), body)
				}
			}
		}

	default:
		// XXX: rewrite with typed handlers to fix 405 errors on all urls?
		status = http.StatusMethodNotAllowed
	}

	if status == 0 {
		status = http.StatusNotFound
	}
	ctx.SetStatusCode(status)

}
