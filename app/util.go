package app

import (
	"bytes"
	"errors"
	"log"
	"net/http"
	"runtime"
	"runtime/pprof"
	"sync/atomic"
	"time"

	"github.com/valyala/fasthttp"
)

var (
	errEmptyInt               = errors.New("empty integer")
	errUnexpectedFirstChar    = errors.New("unexpected first char found. Expecting 0-9")
	errUnexpectedTrailingChar = errors.New("unexpected traling char found. Expecting 0-9")
	errTooLongInt             = errors.New("too long int")
)

var maxIntChars = 10

func parseUint32(b []byte) (uint32, error) {
	n := len(b)
	if n == 0 {
		return 0, errEmptyInt
	}
	var v uint32
	for i := 0; i < n; i++ {
		c := b[i]
		k := c - '0'
		if k > 9 {
			if i == 0 {
				return 0, errUnexpectedFirstChar
			}
			return 0, errUnexpectedTrailingChar
		}
		if i >= maxIntChars {
			return 0, errTooLongInt
		}
		v = 10*v + uint32(k)
	}
	if n != len(b) {
		return 0, errUnexpectedTrailingChar
	}
	return v, nil
}

func (app *Application) RpsWatcher() {
	for {
		time.Sleep(1 * time.Second)
		count := atomic.LoadInt32(&app.countRequests)
		if count > 0 {
			log.Printf("RPS: %6d", count)
			atomic.SwapInt32(&app.countRequests, 0)
		}
	}
}

func GetPprof(ctx *fasthttp.RequestCtx, entity []byte) int {
	if bytes.Equal(entity, []byte("pprof")) {
		t, err := time.ParseDuration(string(ctx.QueryArgs().Peek("t")))
		if err != nil {
			t = 30 * time.Second
		}
		if err := pprof.StartCPUProfile(ctx); err != nil {
			log.Print("could not start CPU profile: ", err)
			return http.StatusInternalServerError
		}
		time.Sleep(t)
		pprof.StopCPUProfile()
		return http.StatusOK
	}
	if bytes.Equal(entity, []byte("pprof_mem")) {
		runtime.GC() // get up-to-date statistics
		if err := pprof.WriteHeapProfile(ctx); err != nil {
			log.Fatal("could not write memory profile: ", err)
		}
		return http.StatusOK
	}
	return http.StatusNotFound
}
