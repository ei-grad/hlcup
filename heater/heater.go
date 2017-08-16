package heater

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/valyala/fasthttp"

	"github.com/ei-grad/hlcup/models"
)

type heater struct {
	baseURL, fileName string
	wg                sync.WaitGroup
	nWorkers          int
}

func RunHeater(baseURL, fileName string, nWorkers int) {
	l := &heater{
		baseURL:  baseURL,
		fileName: fileName,
		nWorkers: nWorkers,
	}
	l.Run()
}

func (l *heater) Run() {

	urls := make(chan string)
	defer close(urls)

	for i := 0; i < l.nWorkers; i++ {
		go l.worker(urls)
	}

	// Open a zip archive for reading.
	r, err := zip.OpenReader(l.fileName)
	if err != nil {
		log.Fatal(err)
	}
	defer r.Close()

	log.Printf("heater: starting")

	t0 := time.Now()

	for _, f := range r.File {
		l.wg.Add(1)
		go l.loadFile(f, 1, urls)
	}

	l.wg.Wait()

	log.Printf("heater: finished in %s", time.Since(t0))

}

func (l *heater) loadFile(f *zip.File, stage int, urls chan string) {

	defer l.wg.Done()

	rc, err := f.Open()
	if err != nil {
		log.Fatal(err)
	}
	defer rc.Close()

	decoder := json.NewDecoder(rc)

	// read left_bracket token
	token, err := decoder.Token()
	if err != nil {
		log.Fatalf("heater: %s: invalid JSON", f.Name)
	}
	if t, ok := token.(json.Delim); !ok || t.String() != "{" {
		log.Fatalf("heater: %s: expected {, got %v", f.Name, token)
	}

	for decoder.More() {

		// read key
		token, err = decoder.Token()
		if err != nil {
			log.Fatalf("heater: %s: invalid JSON", f.Name)
		}
		key, ok := token.(string)
		if !ok {
			log.Fatalf("heater: %s: expected string, got %v", f.Name, token)
		}

		// read left_brace token
		token, err = decoder.Token()
		if err != nil {
			log.Fatalf("heater: %s: invalid JSON", f.Name)
		}
		if t, ok := token.(json.Delim); !ok || t.String() != "[" {
			log.Fatalf("heater: %s: expected [, got %v", f.Name, token)
		}

		var entityHandler func()

		switch {
		case key == "users":
			entityHandler = l.usersHandler(decoder, urls)
		case key == "locations":
			entityHandler = l.locationsHandler(decoder, urls)
		case key == "visits":
			entityHandler = l.visitsHandler(decoder, urls)
		default:
			return
		}

		for decoder.More() {
			entityHandler()
		}
	}

}

func (l *heater) usersHandler(decoder *json.Decoder, urls chan string) func() {
	return func() {
		var v models.User
		err := decoder.Decode(&v)
		if err != nil {
			log.Fatalf("heater: bad JSON: %s", err)
		}
		l.wg.Add(2)
		urls <- fmt.Sprintf("%s/users/%d", l.baseURL, v.ID)
		urls <- fmt.Sprintf("%s/users/%d/visits", l.baseURL, v.ID)
	}
}

func (l *heater) locationsHandler(decoder *json.Decoder, urls chan string) func() {
	return func() {
		var v models.Location
		err := decoder.Decode(&v)
		if err != nil {
			log.Fatalf("heater: bad JSON: %s", err)
		}
		l.wg.Add(2)
		urls <- fmt.Sprintf("%s/locations/%d", l.baseURL, v.ID)
		urls <- fmt.Sprintf("%s/locations/%d/avg", l.baseURL, v.ID)
	}
}

func (l *heater) visitsHandler(decoder *json.Decoder, urls chan string) func() {
	return func() {
		var v models.Visit
		err := decoder.Decode(&v)
		if err != nil {
			log.Fatalf("heater: bad JSON: %s", err)
		}
		l.wg.Add(1)
		urls <- fmt.Sprintf("%s/visits/%d", l.baseURL, v.ID)
	}
}

func (l *heater) worker(urls chan string) {
	for i := range urls {
		l.sendGet(i)
	}
}

func (l *heater) sendGet(url string) {

	defer l.wg.Done()

	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	req.Header.SetMethod("GET")
	req.SetRequestURI(url)

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	err := fasthttp.Do(req, resp)

	if err != nil {
		log.Fatalf("heater: can't send request: %s", err)
	}

	if resp.StatusCode() != 200 {
		log.Fatalf("heater: %s got non-200 response:\n%s", url, resp)
	}

}
