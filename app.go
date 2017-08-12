package main

import (
	"bytes"
	"fmt"
	"log"
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
	app.cache, err = bigcache.NewBigCache(bigcache.DefaultConfig(10 * time.Minute))
	if err != nil {
		log.Fatalf("Can't create bigcache: %s", err)
	}

	app.db = db.New()

	return app

}

func (app Application) requestHandler(ctx *fasthttp.RequestCtx) {

	ctx.SetContentType(strApplicationJSON)

	parts := bytes.SplitN(ctx.Path(), []byte("/"), 3)

	switch string(ctx.Method()) {

	case strGet:
		var resp []byte
		switch len(parts) {
		case 3:
			entity := string(parts[1])
			id, err := strconv.Atoi(string(parts[2]))
			if err != nil {
				break
			}
			if v, err := app.cache.Get(string(ctx.Path())); err == nil {
				resp = v
				break
			}
			switch entity {
			case strUsers:
				v := app.db.GetUser(uint32(id))
				if !v.Valid {
					break
				}
				resp, err = v.MarshalJSON()
			case strLocations:
				v := app.db.GetLocation(uint32(id))
				if !v.Valid {
					break
				}
				resp, err = v.MarshalJSON()
			case strVisits:
				v := app.db.GetVisit(uint32(id))
				if !v.Valid {
					break
				}
				resp, err = v.MarshalJSON()
			}
			if err != nil {
				ctx.Logger().Printf(err.Error())
				ctx.SetStatusCode(500)
				return
			}
			if err = app.cache.Set(string(ctx.Path()), resp); err != nil {
				ctx.Logger().Printf(err.Error())
			}
		case 4:
			entity := string(parts[1])
			_, err := strconv.Atoi(string(parts[2]))
			if err != nil {
				break
			}
			tail := string(parts[3])
			if entity == "users" && tail == "visits" {
				// TODO: implement /users/<id>/visits
			} else if entity == "locations" && tail == "avg" {
				// TODO: implement /locations/<id>/avg
			}
		}
		if resp != nil {
			ctx.Write(resp)
		} else {
			ctx.SetStatusCode(404)
		}

	case strPost:

		// just {} response for POST requests, and Connection:close, yeah
		ctx.SetConnectionClose()
		ctx.Write([]byte("{}"))

		if len(parts) != 3 {
			ctx.SetStatusCode(404)
		} else if string(parts[2]) == "new" {
			entity := string(parts[1])
			body := ctx.PostBody()
			switch entity {
			case strUsers:
				var v models.User
				if err := ffjson.Unmarshal(body, &v); err != nil {
					ctx.SetStatusCode(400)
					ctx.Logger().Printf(err.Error())
					return
				}
				if err := v.Validate(); err != nil {
					ctx.SetStatusCode(400)
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
					ctx.SetStatusCode(400)
					return
				}
				if err := v.Validate(); err != nil {
					ctx.Logger().Printf(err.Error())
					ctx.SetStatusCode(400)
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
					ctx.SetStatusCode(400)
					ctx.Logger().Printf(err.Error())
					return
				}
				if err := v.Validate(); err != nil {
					ctx.SetStatusCode(400)
					ctx.Logger().Printf(err.Error())
					return
				}
				// XXX: what if it already exists?
				if err := app.db.AddVisit(v); err != nil {
					ctx.SetStatusCode(400)
					ctx.Logger().Printf(err.Error())
					return
				}
				bodyCopy := make([]byte, len(body))
				copy(bodyCopy, body)
				if err := app.cache.Set(fmt.Sprintf("/visits/%d", v.ID), bodyCopy); err != nil {
					ctx.Logger().Printf(err.Error())
				}
			default:
				ctx.SetStatusCode(404)
			}
		} else {
			// TODO: implement updating
			//entity := string(parts[1])
			//id, err := strconv.Atoi(string(parts[2]))
			//if err != nil {
			//	ctx.SetStatusCode(404)
			//	break
			//}
			ctx.SetStatusCode(404)
		}

	default:
		ctx.SetStatusCode(405)
	}

	//fmt.Fprintf(ctx, "Query string is %q\n", ctx.QueryArgs())

}
