package main

import (
	"flag"
	"log"
	"syscall"
	"time"

	"github.com/valyala/fasthttp"

	"github.com/ei-grad/hlcup/app"
	"github.com/ei-grad/hlcup/loader"
)

func main() {

	accessLog := flag.Bool("v", false, "show access log")
	address := flag.String("b", ":80", "bind address")
	loaderBaseURL := flag.String("url", "http://localhost", "base URL (for loader)")
	dataFileName := flag.String("data", "/tmp/data/data.zip", "data file name")
	loaderWorkers := flag.Int("loader-workers", 8, "number of parallel requests while loading data")

	flag.Parse()

	h := app.NewApplication().RequestHandler

	if *accessLog {
		h = accessLogHandler(h)
	}

	syscall.Mlockall(syscall.MCL_CURRENT | syscall.MCL_FUTURE)

	go func() {
		time.Sleep(1 * time.Second)
		loader.LoadData(*loaderBaseURL, *dataFileName, *loaderWorkers)
	}()

	if err := fasthttp.ListenAndServe(*address, h); err != nil {
		log.Fatalf("Error in ListenAndServe: %s", err)
	}
}
