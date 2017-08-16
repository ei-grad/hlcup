package main

import (
	"log"
	"os"
	"runtime"
	"time"
)

func init() {
	var memstats runtime.MemStats
	if os.Getenv("DEBUG_MEMSTATS") != "1" {
		return
	}
	go func() {
		for {
			runtime.ReadMemStats(&memstats)
			log.Printf("memstats: %+v", memstats)
			time.Sleep(1 * time.Second)
		}
	}()
}
