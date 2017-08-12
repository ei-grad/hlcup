package main

import (
	"flag"
	"log"

	"github.com/valyala/fasthttp"
)

func main() {

	accessLog := flag.Bool("v", false, "show access log")
	address := flag.String("b", ":80", "bind address")
	loaderBaseURL := flag.String("url", "http://localhost", "base URL (for loader)")
	dataFileName := flag.String("data", "/tmp/data/data.zip", "data file name")

	flag.Parse()

	app := NewApplication()

	h := app.requestHandler

	if *accessLog {
		h = accessLogHandler(h)
	}

	loader := &Loader{
		baseURL:  *loaderBaseURL,
		fileName: *dataFileName,
	}

	go loader.LoadData()

	if err := fasthttp.ListenAndServe(*address, h); err != nil {
		log.Fatalf("Error in ListenAndServe: %s", err)
	}
}
