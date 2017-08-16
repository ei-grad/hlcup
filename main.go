package main

import (
	"flag"
	"log"
	"os"
	"runtime"
	"syscall"
	"time"

	"github.com/valyala/fasthttp"

	"github.com/ei-grad/hlcup/app"
	"github.com/ei-grad/hlcup/loader"
)

var appVersion, appBuildDate string

func main() {

	log.Printf("HighLoad Cup solution by Andrew Grigorev <andrew@ei-grad.ru>")
	log.Printf("Version %s built %s, %s", appVersion, appBuildDate, runtime.Version())

	accessLog := flag.Bool("v", false, "show access log")
	address := flag.String("b", ":80", "bind address")
	loaderBaseURL := flag.String("url", "http://localhost", "base URL (for loader)")
	dataFileName := flag.String("data", "/tmp/data/data.zip", "data file name")
	loaderWorkers := flag.Int("loader-workers", 8, "number of parallel requests while loading data")

	flag.Parse()

	rlimit()
	memstats()

	if os.Getenv("RUN_TOP") == "1" {
		go top()
	}

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
