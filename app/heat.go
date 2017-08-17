package app

import (
	"log"

	"github.com/valyala/fasthttp"

	"github.com/ei-grad/hlcup/entities"
)

func (app *Application) UseHeat(heat bool) {
	app.heat = func(entity entities.Entity, id uint32) {

		buf := fasthttp.AcquireByteBuffer()
		defer fasthttp.ReleaseByteBuffer(buf)
		if status := app.GetEntity(buf, entity, id); status != 200 {
			log.Fatalf("heat: got non-200 response: GET /%s/%d %d", entities.GetEntityRoute(entity), id, status)
		}

		args := fasthttp.AcquireArgs()
		defer fasthttp.ReleaseArgs(args)

		switch entity {
		case entities.User:
			if status := app.GetUserVisits(buf, id, args); status != 200 {
				log.Fatalf("heat: got non-200 response: GET /users/%d/visits %d", id, status)
			}
		case entities.Location:
			if status := app.GetLocationAvg(buf, id, args); status != 200 {
				log.Fatalf("heat: got non-200 response: GET /locations/%d/marks %d", id, status)
			}
		}
	}
}
