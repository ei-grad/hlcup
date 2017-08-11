package main

import (
	"bytes"
	"flag"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/allegro/bigcache"
	"github.com/pquerna/ffjson/ffjson"
	"github.com/valyala/fasthttp"

	"github.com/ei-grad/hlcup/entities"
)

func main() {

	accessLog := flag.Bool("v", false, "show access log")
	address := flag.String("b", ":80", "bind address")

	flag.Parse()

	h := requestHandler

	if *accessLog {
		h = accessLogHandler(h)
	}

	go loadData()
	if err := fasthttp.ListenAndServe(*address, h); err != nil {
		log.Fatalf("Error in ListenAndServe: %s", err)
	}
}

var users *entities.UserMap
var locations *entities.LocationMap
var visits *entities.VisitMap

var cache *bigcache.BigCache

func init() {
	users = entities.NewUserMap(509)
	locations = entities.NewLocationMap(509)
	visits = entities.NewVisitMap(509)

	var err error
	cache, err = bigcache.NewBigCache(bigcache.DefaultConfig(10 * time.Minute))
	if err != nil {
		log.Fatalf("Can't create bigcache: %s", err)
	}
}

func accessLogHandler(handler fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		t0 := time.Now()
		handler(ctx)
		log.Printf(
			"\"%s\" %d %f",
			strings.Split(ctx.Request.Header.String(), "\r\n")[0],
			ctx.Response.StatusCode(),
			time.Now().Sub(t0).Seconds(),
		)
	}
}

func requestHandler(ctx *fasthttp.RequestCtx) {

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
			if v, err := cache.Get(string(ctx.Path())); err == nil {
				resp = v
				break
			}
			switch entity {
			case strUsers:
				v := users.Get(uint32(id))
				if !v.Valid {
					break
				}
				resp, err = v.MarshalJSON()
			case strLocations:
				v := locations.Get(uint32(id))
				if !v.Valid {
					break
				}
				resp, err = v.MarshalJSON()
			case strVisits:
				v := visits.Get(uint32(id))
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
			if err = cache.Set(string(ctx.Path()), resp); err != nil {
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
		if len(parts) != 3 {
			ctx.SetStatusCode(404)
		} else if string(parts[2]) == "new" {
			entity := string(parts[1])
			switch entity {
			case strUsers:
				var v entities.User
				body := ctx.PostBody()
				if err := ffjson.Unmarshal(body, &v); err != nil {
					ctx.SetStatusCode(400)
					return
				}
				v.Validate()
				// XXX: what if it already exists?
				bodyCopy := make([]byte, len(body))
				copy(bodyCopy, body)
				users.Set(v.ID, v)
			case strLocations:
				var v entities.Location
				body := ctx.PostBody()
				if err := ffjson.Unmarshal(body, &v); err != nil {
					ctx.SetStatusCode(400)
					return
				}
				v.Validate()
				// XXX: what if it already exists?
				bodyCopy := make([]byte, len(body))
				copy(bodyCopy, body)
				locations.Set(v.ID, v)
			case strVisits:
				var v entities.Visit
				body := ctx.PostBody()
				if err := ffjson.Unmarshal(body, &v); err != nil {
					ctx.SetStatusCode(400)
					return
				}
				v.Validate()
				// XXX: what if it already exists?
				bodyCopy := make([]byte, len(body))
				copy(bodyCopy, body)
				visits.Set(v.ID, v)
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

	//fmt.Fprintf(ctx, "Request method is %q\n", ctx.Method())
	//fmt.Fprintf(ctx, "RequestURI is %q\n", ctx.RequestURI())
	//fmt.Fprintf(ctx, "Requested path is %q\n", ctx.Path())
	//fmt.Fprintf(ctx, "Host is %q\n", ctx.Host())
	//fmt.Fprintf(ctx, "Query string is %q\n", ctx.QueryArgs())
	//fmt.Fprintf(ctx, "User-Agent is %q\n", ctx.UserAgent())
	//fmt.Fprintf(ctx, "Connection has been established at %s\n", ctx.ConnTime())
	//fmt.Fprintf(ctx, "Request has been started at %s\n", ctx.Time())
	//fmt.Fprintf(ctx, "Serial request number for the current connection is %d\n", ctx.ConnRequestNum())
	//fmt.Fprintf(ctx, "Your ip is %q\n\n", ctx.RemoteIP())
	//fmt.Fprintf(ctx, "Raw request is:\n---CUT---\n%s\n---CUT---", &ctx.Request)

}
