package main

import (
	"flag"
	"log"
	"os"
	"runtime"
	"runtime/pprof"
	"syscall"
	"time"

	"github.com/valyala/fasthttp"

	"github.com/ei-grad/hlcup/app"
)

var appVersion, appBuildDate string

func main() {

	log.Printf("HighLoad Cup solution by Andrew Grigorev <andrew@ei-grad.ru>")
	log.Printf("Version %s built %s, %s", appVersion, appBuildDate, runtime.Version())

	var (
		accessLog    = flag.Bool("v", false, "show access log")
		address      = flag.String("b", ":80", "bind address")
		dataFileName = flag.String("data", "/tmp/data/data.zip", "data file name")
		cpuprofile   = flag.String("cpuprofile", "", "write cpu profile `file`")
		memprofile   = flag.String("memprofile", "", "write memory profile to `file`")
		useHeat      = flag.Bool("heat", false, "heat GET requests on POST")
	)

	flag.Parse()

	rlimit()
	memstats()

	if os.Getenv("RUN_TOP") == "1" {
		go top()
	}

	app := app.NewApplication()
	app.UseHeat(*useHeat)

	h := app.RequestHandler

	if *accessLog {
		h = accessLogHandler(h)
	}

	syscall.Mlockall(syscall.MCL_CURRENT | syscall.MCL_FUTURE)

	// goroutine to load data and profile cpu and mem
	go func() {

		time.Sleep(1 * time.Second)

		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal("could not create CPU profile: ", err)
		}
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatal("could not start CPU profile: ", err)
		}
		defer pprof.StopCPUProfile()

		app.LoadData(*dataFileName)

		if *memprofile != "" {
			f, err := os.Create(*memprofile)
			if err != nil {
				log.Fatal("could not create memory profile: ", err)
			}
			runtime.GC() // get up-to-date statistics
			if err := pprof.WriteHeapProfile(f); err != nil {
				log.Fatal("could not write memory profile: ", err)
			}
			f.Close()
		}

	}()

	if err := fasthttp.ListenAndServe(*address, h); err != nil {
		log.Fatalf("Error in ListenAndServe: %s", err)
	}
}
