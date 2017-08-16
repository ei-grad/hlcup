package app

import (
	"log"

	"github.com/valyala/fasthttp"
)

func (app *Application) heat(entity string, id uint32) {
	buf := fasthttp.AcquireByteBuffer()
	defer fasthttp.ReleaseByteBuffer(buf)
	if status := app.getEntity(buf, entity, id); status != 200 {
		log.Printf("heat: got non-200 response: GET /%s/%d", entity, id)
	}
}
