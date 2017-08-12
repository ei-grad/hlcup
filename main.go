package main

import (
	"flag"
	"log"

	"github.com/valyala/fasthttp"
)

func main() {

	accessLog := flag.Bool("v", false, "show access log")
	address := flag.String("b", ":80", "bind address")

	flag.Parse()

	app := NewApplication()

	h := app.requestHandler

	if *accessLog {
		h = accessLogHandler(h)
	}

	go loadData()
	if err := fasthttp.ListenAndServe(*address, h); err != nil {
		log.Fatalf("Error in ListenAndServe: %s", err)
	}
}
