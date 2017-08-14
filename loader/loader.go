package loader

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/pquerna/ffjson/ffjson"
	"github.com/valyala/fasthttp"

	"github.com/ei-grad/hlcup/models"
)

type task struct {
	url  string
	body []byte
}

type loader struct {
	baseURL, fileName string
	wg                sync.WaitGroup
	nWorkers          int
	countUsers        int64
	countLocations    int64
	countVisits       int64
}

func LoadData(baseURL, fileName string, nWorkers int) {
	l := &loader{
		baseURL:  baseURL,
		fileName: fileName,
		nWorkers: nWorkers,
	}
	l.loadData()
}

func (l *loader) loadData() {

	tasks := make(chan task)
	defer close(tasks)

	for i := 0; i < l.nWorkers; i++ {
		go l.worker(tasks)
	}

	// Open a zip archive for reading.
	r, err := zip.OpenReader(l.fileName)
	if err != nil {
		log.Fatal(err)
	}
	defer r.Close()

	log.Printf("loader: starting")

	t0 := time.Now()

	for _, f := range r.File {
		l.wg.Add(1)
		go l.loadFile(f, 1, tasks)
	}

	l.wg.Wait()

	t1 := time.Now()
	log.Printf("loader: stage 1 finished in %s", t1.Sub(t0))

	for _, f := range r.File {
		l.wg.Add(1)
		go l.loadFile(f, 2, tasks)
	}

	l.wg.Wait()

	t2 := time.Now()
	log.Printf("loader: stage 2 finished in %s", t2.Sub(t1))
	log.Printf("loader: load finished in %s", t2.Sub(t0))
	log.Printf("loader: loaded %d users, %d locations, %d visits",
		l.countUsers, l.countLocations, l.countVisits)

}

func (l *loader) loadFile(f *zip.File, stage int, tasks chan task) {

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
		log.Fatalf("loader: %s: invalid JSON", f.Name)
	}
	if t, ok := token.(json.Delim); !ok || t.String() != "{" {
		log.Fatalf("loader: %s: expected {, got %v", f.Name, token)
	}

	for decoder.More() {

		// read key
		token, err = decoder.Token()
		if err != nil {
			log.Fatalf("loader: %s: invalid JSON", f.Name)
		}
		key, ok := token.(string)
		if !ok {
			log.Fatalf("loader: %s: expected string, got %v", f.Name, token)
		}

		// read left_brace token
		token, err = decoder.Token()
		if err != nil {
			log.Fatalf("loader: %s: invalid JSON", f.Name)
		}
		if t, ok := token.(json.Delim); !ok || t.String() != "[" {
			log.Fatalf("loader: %s: expected [, got %v", f.Name, token)
		}

		type Entity interface {
			MarshalJSON() ([]byte, error)
			UnmarshalJSON([]byte) error
		}

		var v Entity
		var constructor func() Entity

		switch {
		case key == "users" && stage == 1:
			constructor = func() Entity {
				atomic.AddInt64(&l.countUsers, 1)
				return &models.User{}
			}
		case key == "locations" && stage == 1:
			constructor = func() Entity {
				atomic.AddInt64(&l.countLocations, 1)
				return &models.Location{}
			}
		case key == "visits" && stage == 2:
			constructor = func() Entity {
				atomic.AddInt64(&l.countVisits, 1)
				return &models.Visit{}
			}
		default:
			log.Printf("loader: %s: unknown section '%s', ignoring the remaining contents", f.Name, key)
			return
		}

		log.Printf("loader: %s: loading %s", f.Name, key)

		for decoder.More() {
			v = constructor()
			err := decoder.Decode(&v)
			if err != nil {
				log.Fatalf("loader: bad JSON: %s", err)
			}
			body, err := v.MarshalJSON()
			if err != nil {
				log.Fatalf("loader: can't encode %+v back: %s", v, err)
			}
			l.wg.Add(1)
			tasks <- task{
				url:  fmt.Sprintf("%s/%s/new", l.baseURL, key),
				body: body,
			}
		}
	}

}

func (l *loader) worker(tasks chan task) {
	for i := range tasks {
		l.sendPost(i.url, i.body)
		l.wg.Done()
	}
}

func (l *loader) sendPost(url string, body []byte) {

	defer ffjson.Pool(body)

	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	req.Header.SetMethod("POST")
	req.SetRequestURI(url)
	req.Header.SetContentType("application/json")
	req.SetBody(body)

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	err := fasthttp.Do(req, resp)

	if err != nil {
		log.Fatalf("loader: can't send request: %s", err)
	}

	if resp.StatusCode() != 200 {
		log.Fatalf("loader: LOAD FAILED! Got non-200 response:\n%s", resp)
	}

}
