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
			"%d %s",
			ctx.Response.StatusCode(),
			time.Since(t0),
		)
	}
}
