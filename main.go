package main

import (
	"bytes"
	"flag"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/pquerna/ffjson/ffjson"
	"github.com/valyala/fasthttp"

	"github.com/ei-grad/hlcup/entities"
	"github.com/ei-grad/hlcup/maps"
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

var users *maps.UserMap
var locations *maps.LocationMap
var visits *maps.VisitMap

func init() {
	users = maps.NewUserMap(509)
	locations = maps.NewLocationMap(509)
	visits = maps.NewVisitMap(509)
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

	case "GET":
		var resp []byte
		switch len(parts) {
		case 3:
			entity := string(parts[1])
			id := string(parts[2])
			switch entity {
			case strUsers:
				resp = users.Get(id).JSON
			case strLocations:
				resp = locations.Get(id).JSON
			case strVisits:
				resp = visits.Get(id).JSON
			}
		case 4:
			entity := string(parts[1])
			//id := string(parts[2])
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

	case "POST":
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
				// XXX: what if it already exists?
				users.Set(strconv.Itoa(int(v.ID)), maps.User{
					Parsed: v,
					JSON:   body,
				})
			case strLocations:
				var v entities.Location
				body := ctx.PostBody()
				if err := ffjson.Unmarshal(body, &v); err != nil {
					ctx.SetStatusCode(400)
					return
				}
				// XXX: what if it already exists?
				locations.Set(strconv.Itoa(int(v.ID)), maps.Location{
					Parsed: v,
					JSON:   body,
				})
			case strVisits:
				var v entities.Visit
				body := ctx.PostBody()
				if err := ffjson.Unmarshal(body, &v); err != nil {
					ctx.SetStatusCode(400)
					return
				}
				// XXX: what if it already exists?
				visits.Set(strconv.Itoa(int(v.ID)), maps.Visit{
					Parsed: v,
					JSON:   body,
				})
			default:
				ctx.SetStatusCode(404)
			}
		} else {
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
