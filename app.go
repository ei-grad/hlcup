package main

import (
	"bytes"
	"net/http"
	"strconv"

	"github.com/valyala/fasthttp"

	"github.com/ei-grad/hlcup/db"
)

// Application implements application logic
type Application struct {
	db *db.DB
}

// NewApplication creates new Application
func NewApplication() (app Application) {
	app.db = db.New()
	return
}

func parseUint32(s []byte) (uint32, error) {
	parsed, err := strconv.ParseUint(string(s), 10, 32)
	if err != nil {
		return 0, err
	}
	return uint32(parsed), nil
}

// RequestHandler contains implementation of all routes and most of application
// logic
func (app Application) RequestHandler(ctx *fasthttp.RequestCtx) {

	ctx.SetContentType(strApplicationJSON)

	parts := bytes.SplitN(ctx.Path(), []byte("/"), 4)

	var (
		id     uint32
		status int
		err    error
	)

	switch string(ctx.Method()) {

	case strGet:
		switch len(parts) {
		case 3:
			if string(parts[2]) == "new" {
				status = http.StatusMethodNotAllowed
			} else if id, err = parseUint32(parts[2]); err != nil {
				status = http.StatusNotFound
			} else {
				status = app.getEntity(ctx, string(parts[1]), id)
			}
		case 4:
			if id, err = parseUint32(parts[2]); err != nil {
				status = http.StatusNotFound
			} else {
				entity := string(parts[1])
				tail := string(parts[3])
				switch {
				case entity == "users" && tail == "visits":
					status = app.getUserVisits(ctx, id)
				case entity == "locations" && tail == "avg":
					status = app.getLocationAvg(ctx, id)
				case entity == "locations" && tail == "marks":
					status = app.getLocationMarks(ctx, id)
				default:
					status = http.StatusNotFound
				}
			}
		default:
			status = http.StatusNotFound
		}
	case strPost:

		// To fix the "Empty response" error in yandex-tank logs we have to send
		// "Connection: close" for POST requests.
		ctx.SetConnectionClose()

		// Also, check system expects a {} in the response body
		ctx.Write([]byte("{}"))

		if len(parts) != 3 {
			status = http.StatusNotFound
			break
		}

		body := ctx.PostBody()

		// XXX: some email's could contain the null string... but hopefully - not :-)
		//if bytes.Contains(body, []byte("null")) {
		//	ctx.Logger().Printf("found null value: %s", body)
		//	status = http.StatusBadRequest
		//	break
		//}

		if string(parts[2]) == "new" {
			status = app.postEntityNew(ctx, string(parts[1]), body)
		} else {
			if id, err = parseUint32(parts[2]); err != nil {
				status = http.StatusNotFound
				break
			}
			status = app.postEntity(ctx, string(parts[1]), id, body)
		}

	default:
		// XXX: rewrite with typed handlers to fix 405 errors on all urls?
		status = http.StatusMethodNotAllowed
	}

	ctx.SetStatusCode(status)

}
