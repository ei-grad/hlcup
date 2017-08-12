package main

import (
	"github.com/valyala/fasthttp"
	"time"
)

func accessLogHandler(handler fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		t0 := time.Now()
		handler(ctx)
		ctx.Logger().Printf(
			"%d %03.3fms",
			ctx.Response.StatusCode(),
			time.Now().Sub(t0).Seconds()/1000.,
		)
	}
}
