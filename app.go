package main

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/allegro/bigcache"
	"github.com/pquerna/ffjson/ffjson"
	"github.com/valyala/fasthttp"

	"github.com/ei-grad/hlcup/db"
	"github.com/ei-grad/hlcup/models"
)

type Application struct {
	cache *bigcache.BigCache
	db    *db.DB
}

func NewApplication() Application {

	var app Application

	var err error

	config := bigcache.Config{
		// number of shards (must be a power of 2)
		Shards: 1024,
		// time after which entry can be evicted
		LifeWindow: 10 * time.Minute,
		// rps * lifeWindow, used only in initial memory allocation
		MaxEntriesInWindow: 100 * 10 * 60,
		// max entry size in bytes, used only in initial memory allocation
		MaxEntrySize: 500,
		// prints information about additional memory allocation
		Verbose: true,
		// cache will not allocate more memory than this limit, value in MB
		// if value is reached then the oldest entries can be overridden for the new ones
		// 0 value means no size limit
		HardMaxCacheSize: 1024,
		// callback fired when the oldest entry is removed because of its
		// expiration time or no space left for the new entry. Default value is nil which
		// means no callback and it prevents from unwrapping the oldest entry.
		OnRemove: nil,
	}

	app.cache, err = bigcache.NewBigCache(config)
	if err != nil {
		log.Fatalf("Can't create bigcache: %s", err)
	}

	log.Printf("Cache initialized.")

	app.db = db.New()

	return app

}

func (app Application) requestHandler(ctx *fasthttp.RequestCtx) {

	ctx.SetContentType(strApplicationJSON)

	parts := bytes.SplitN(ctx.Path(), []byte("/"), 4)

	switch string(ctx.Method()) {

	case strGet:

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

			if v, err := app.cache.Get(string(ctx.Path())); err == nil {
				// response from cache
				ctx.Write(v)
				return
			}

			switch entity {

			case strUsers:
				v := app.db.GetUser(id)
				if !v.Valid {
					// 404 - user with given ID doesn't exist
					ctx.SetStatusCode(http.StatusNotFound)
					return
				}
				resp, err = v.MarshalJSON()

			case strLocations:
				v := app.db.GetLocation(id)
				if !v.Valid {
					// 404 - location with given ID doesn't exist
					ctx.SetStatusCode(http.StatusNotFound)
					return
				}
				resp, err = v.MarshalJSON()

			case strVisits:
				v := app.db.GetVisit(id)
				if !v.Valid {
					// 404 - visit with given ID doesn't exist
					ctx.SetStatusCode(http.StatusNotFound)
					return
				}
				resp, err = v.MarshalJSON()

			}

			if err != nil {
				// ffjson marshal failed, shouldn't happen
				panic(err)
			}

			ctx.Write(resp)

			if err = app.cache.Set(string(ctx.Path()), resp); err != nil {
				// bigcache set failed, shouldn't happen
				panic(err)
			}

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

			if entity == "users" && tail == "visits" {

				visits := app.db.GetUserVisits(id)
				if visits == nil {
					// 404 - user have no visits
					ctx.SetStatusCode(http.StatusNotFound)
					return
				}
				ctx.WriteString(`{"visits":[`)
				tmp, _ := visits[0].MarshalJSON()
				ctx.Write(tmp)
				for _, i := range visits[1:] {
					// TODO: implement /users/<id>/visits filters
					ctx.WriteString(",")
					tmp, _ = i.MarshalJSON()
					ctx.Write(tmp)
				}
				ctx.WriteString("]}")
				return

			} else if entity == "locations" && tail == "avg" {

				marks := app.db.GetLocationMarks(id)
				if marks == nil {
					// 404 - no marks for specified location
					ctx.SetStatusCode(http.StatusNotFound)
					return
				}
				var sum, count int
				for _, i := range marks {
					// TODO: implement /locations/<id>/avg filters
					sum = sum + int(i.Mark)
					count = count + 1
				}
				ctx.WriteString(fmt.Sprintf(`{"avg": %.5f}`, float64(sum)/float64(count)))
				return

			}

		}

		ctx.SetStatusCode(http.StatusNotFound)

	case strPost:

		// just {} response for POST requests, and Connection:close, yeah
		ctx.SetConnectionClose()
		ctx.Write([]byte("{}"))

		if len(parts) != 3 {
			ctx.SetStatusCode(http.StatusNotFound)
			return
		}

		if string(parts[2]) == "new" {
			entity := string(parts[1])
			body := ctx.PostBody()
			switch entity {
			case strUsers:
				var v models.User
				if err := ffjson.Unmarshal(body, &v); err != nil {
					ctx.SetStatusCode(http.StatusBadRequest)
					ctx.Logger().Printf(err.Error())
					return
				}
				if err := v.Validate(); err != nil {
					ctx.SetStatusCode(http.StatusBadRequest)
					ctx.Logger().Printf(err.Error())
					return
				}
				// XXX: what if it already exists?
				bodyCopy := make([]byte, len(body))
				copy(bodyCopy, body)
				app.db.AddUser(v)
				if err := app.cache.Set(fmt.Sprintf("/users/%d", v.ID), bodyCopy); err != nil {
					ctx.Logger().Printf(err.Error())
				}
			case strLocations:
				var v models.Location
				if err := ffjson.Unmarshal(body, &v); err != nil {
					ctx.Logger().Printf(err.Error())
					ctx.SetStatusCode(http.StatusBadRequest)
					return
				}
				if err := v.Validate(); err != nil {
					ctx.Logger().Printf(err.Error())
					ctx.SetStatusCode(http.StatusBadRequest)
					return
				}
				// XXX: what if it already exists?
				bodyCopy := make([]byte, len(body))
				copy(bodyCopy, body)
				app.db.AddLocation(v)
				if err := app.cache.Set(fmt.Sprintf("/locations/%d", v.ID), bodyCopy); err != nil {
					ctx.Logger().Printf(err.Error())
				}
			case strVisits:
				var v models.Visit
				if err := ffjson.Unmarshal(body, &v); err != nil {
					ctx.SetStatusCode(http.StatusBadRequest)
					ctx.Logger().Printf(err.Error())
					return
				}
				if err := v.Validate(); err != nil {
					ctx.SetStatusCode(http.StatusBadRequest)
					ctx.Logger().Printf(err.Error())
					return
				}
				// XXX: what if it already exists?
				if err := app.db.AddVisit(v); err != nil {
					ctx.SetStatusCode(http.StatusBadRequest)
					ctx.Logger().Printf(err.Error())
					return
				}
				bodyCopy := make([]byte, len(body))
				copy(bodyCopy, body)
				if err := app.cache.Set(fmt.Sprintf("/visits/%d", v.ID), bodyCopy); err != nil {
					ctx.Logger().Printf(err.Error())
				}
			default:
				ctx.SetStatusCode(http.StatusNotFound)
			}
		} else {
			// TODO: implement updating
			ctx.SetStatusCode(http.StatusNotFound)
			return
		}

	default:
		ctx.SetStatusCode(http.StatusMethodNotAllowed)
		return
	}

}
