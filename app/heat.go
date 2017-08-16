package app

import (
	"log"

	"github.com/valyala/fasthttp"
)

func (app *Application) heat(entity string, id uint32) {

	buf := fasthttp.AcquireByteBuffer()
	defer fasthttp.ReleaseByteBuffer(buf)
	if status := app.GetEntity(buf, entity, id); status != 200 {
		log.Fatalf("heat: got non-200 response: GET /%s/%d", entity, id)
	}

	args := fasthttp.AcquireArgs()
	defer fasthttp.ReleaseArgs(args)

	switch entity {
	case strUsers:
		if status := app.GetUserVisits(buf, id, args); status != 200 {
			log.Fatalf("heat: got non-200 response: GET /users/%d/visits", id)
		}
	case strLocations:
		if status := app.GetLocationAvg(buf, id, args); status != 200 {
			log.Fatalf("heat: got non-200 response: GET /locations/%d/marks", id)
		}
	}
}
