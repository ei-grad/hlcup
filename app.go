package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/valyala/fasthttp"

	"github.com/ei-grad/hlcup/db"
	"github.com/ei-grad/hlcup/models"
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

// RequestHandler contains implementation of all routes and most of application
// logic
func (app Application) RequestHandler(ctx *fasthttp.RequestCtx) {

	ctx.SetContentType(strApplicationJSON)

	parts := bytes.SplitN(ctx.Path(), []byte("/"), 4)

	switch string(ctx.Method()) {

	case strGet:

		var v interface {
			IsValid() bool
			MarshalJSON() ([]byte, error)
		}

		switch len(parts) {
		case 3:

			var resp []byte
			var id uint32

			entity := string(parts[1])

			id64, err := strconv.ParseUint(string(parts[2]), 10, 32)
			if err != nil {
				// 404 - id is not integer
				ctx.SetStatusCode(http.StatusNotFound)
				return
			}
			id = uint32(id64)

			switch entity {
			case strUsers:
				user := app.db.GetUser(id)
				v = &user
			case strLocations:
				location := app.db.GetLocation(id)
				v = &location
			case strVisits:
				visit := app.db.GetVisit(id)
				v = &visit
			default:
				ctx.SetStatusCode(http.StatusNotFound)
				return
			}

			if !v.IsValid() {
				// 404 - user with given ID doesn't exist
				ctx.SetStatusCode(http.StatusNotFound)
				return
			}

			resp, err = v.MarshalJSON()
			if err != nil {
				// v.MarshalJSON() failed, shouldn't happen
				panic(err)
			}

			ctx.Write(resp)

			return

		case 4:

			entity := string(parts[1])

			id64, err := strconv.ParseUint(string(parts[2]), 10, 32)
			if err != nil {
				// 404 - id is not integer
				ctx.SetStatusCode(http.StatusNotFound)
				return
			}
			id := uint32(id64)

			tail := string(parts[3])

			switch {

			case entity == "users" && tail == "visits":

				if !app.db.GetUser(id).IsValid() {
					ctx.SetStatusCode(http.StatusNotFound)
					return
				}

				filter, err := GetVisitsFilter(ctx.QueryArgs())
				if err != nil {
					ctx.SetStatusCode(http.StatusBadRequest)
					return
				}

				visits := app.db.GetUserVisits(id)
				if visits == nil {
					// user have no visits
					ctx.WriteString(`{"visits":[]}`)
					return
				}

				first := true

				ctx.WriteString(`{"visits":[`)
				visits.M.RLock()
				for _, i := range visits.Visits {
					// TODO: implement /users/<id>/visits filters
					if !filter(i) {
						continue
					}
					if !first {
						ctx.WriteString(",")
					}
					tmp, _ := i.MarshalJSON()
					ctx.Write(tmp)
					first = false
				}
				visits.M.RUnlock()
				ctx.WriteString("]}")
				return

			case entity == "locations" && tail == "avg":

				if !app.db.GetLocation(id).IsValid() {
					ctx.SetStatusCode(http.StatusNotFound)
					return
				}

				filter, err := GetMarksFilter(ctx.QueryArgs())
				if err != nil {
					ctx.SetStatusCode(http.StatusBadRequest)
					return
				}

				var sum, count int
				var avg float64

				marks := app.db.GetLocationMarks(id)
				if marks != nil {
					marks.M.RLock()
					for _, i := range marks.Marks {
						if !filter(i) {
							continue
						}
						sum = sum + int(i.Mark)
						count = count + 1
					}
					marks.M.RUnlock()
				}
				if count == 0 {
					// location have no marks
					avg = 0.
				} else {
					avg = float64(sum) / float64(count)
				}
				ctx.WriteString(fmt.Sprintf(`{"avg": %.5f}`, avg))
				return

			case entity == "locations" && tail == "marks":

				if !app.db.GetLocation(id).IsValid() {
					ctx.SetStatusCode(http.StatusNotFound)
					return
				}

				first := true

				ctx.WriteString(`{"marks":[`)
				marks := app.db.GetLocationMarks(id)
				if marks != nil {
					marks.M.RLock()
					for _, i := range marks.Marks {
						if !first {
							ctx.WriteString(",")
						}
						tmp, _ := i.MarshalJSON()
						ctx.Write(tmp)
						first = false
					}
				}
				marks.M.RUnlock()
				ctx.WriteString("]}")
				return

			default:
				ctx.SetStatusCode(http.StatusNotFound)
				return

			}

		}

	case strPost:

		// To fix the "Empty response" error in yandex-tank logs we have to send
		// "Connection: close" for POST requests.
		ctx.SetConnectionClose()

		// Also, check system expects a {} in the response body
		ctx.Write([]byte("{}"))

		if len(parts) != 3 {
			ctx.SetStatusCode(http.StatusNotFound)
			return
		}

		entity := string(parts[1])

		body := ctx.PostBody()

		var v interface {
			UnmarshalJSON([]byte) error
			Validate() error
		}

		if string(parts[2]) == "new" {

			var saver func() error

			switch entity {
			case strUsers:
				var user models.User
				v = &user
				saver = func() error { return app.db.AddUser(user) }
			case strLocations:
				var location models.Location
				v = &location
				saver = func() error { return app.db.AddLocation(location) }
			case strVisits:
				var visit models.Visit
				v = &visit
				saver = func() error { return app.db.AddVisit(visit) }
			default:
				ctx.SetStatusCode(http.StatusNotFound)
				return
			}

			if err := v.UnmarshalJSON(body); err != nil {
				ctx.SetStatusCode(http.StatusBadRequest)
				ctx.Logger().Printf(err.Error())
				return
			}
			if err := v.Validate(); err != nil {
				ctx.SetStatusCode(http.StatusBadRequest)
				ctx.Logger().Printf("validate failed: %s", err.Error())
				return
			}
			if err := saver(); err != nil {
				ctx.SetStatusCode(http.StatusBadRequest)
				ctx.Logger().Printf("can't add %+v: %s\n%s", v, err.Error(), body)
				return
			}

		} else {

			id64, err := strconv.ParseUint(string(parts[2]), 10, 32)
			if err != nil {
				// 404 - id is not integer
				ctx.SetStatusCode(http.StatusNotFound)
				return
			}
			id := uint32(id64)

			var v interface {
				IsValid() bool
			}

			switch entity {
			case strUsers:
				user := app.db.GetUser(id)
				v = &user
			case strLocations:
				location := app.db.GetLocation(id)
				v = &location
			case strVisits:
				visit := app.db.GetVisit(id)
				v = &visit
			default:
				ctx.SetStatusCode(http.StatusNotFound)
				return
			}

			if !v.IsValid() {
				ctx.SetStatusCode(http.StatusNotFound)
				return
			}

			m := map[string]interface{}{}

			if err := json.Unmarshal(body, &m); err != nil {
				ctx.SetStatusCode(http.StatusBadRequest)
				return
			}

			// TODO: update entity

		}

	default:
		ctx.SetStatusCode(http.StatusMethodNotAllowed)
		return
	}

}
