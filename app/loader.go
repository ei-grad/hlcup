package app

import (
	"archive/zip"
	"encoding/json"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ei-grad/hlcup/models"
)

type counts struct {
	Users     int32
	Locations int32
	Visits    int32
}

func (app *Application) LoadData(fileName string) {
	var (
		wg sync.WaitGroup
	)

	// Open a zip archive for reading.
	r, err := zip.OpenReader(fileName)
	if err != nil {
		log.Fatal(err)
	}
	defer r.Close()

	log.Printf("loader: starting")

	t0 := time.Now()

	var c counts

	for _, f := range r.File {
		wg.Add(1)
		go app.loadFile(&wg, f, &c, 1)
	}

	wg.Wait()

	log.Printf("loader: stage 1 finished in %s", time.Since(t0))

	for _, f := range r.File {
		wg.Add(1)
		go app.loadFile(&wg, f, &c, 2)
	}

	wg.Wait()

	log.Printf("loader: load finished in %s", time.Since(t0))
	log.Printf("loader: loaded %d users, %d locations, %d visits",
		c.Users, c.Locations, c.Visits)

}

func (app *Application) loadFile(wg *sync.WaitGroup, f *zip.File, c *counts, stage int) {

	defer wg.Done()

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

		var handler func()

		switch {
		case key == strUsers:
			if stage != 1 {
				return
			}
			handler = func() {
				var v models.User
				err = decoder.Decode(&v)
				if err != nil {
					log.Fatalf("loader: bad user: %s", err)
				}
				err = app.db.AddUser(v)
				if err != nil {
					log.Fatalf("loader: can't add user %d: %s", v.ID, err)
				}
				atomic.AddInt32(&c.Users, 1)
			}
		case key == strLocations:
			if stage != 1 {
				return
			}
			handler = func() {
				var v models.Location
				err = decoder.Decode(&v)
				if err != nil {
					log.Fatalf("loader: bad location: %s", err)
				}
				err = app.db.AddLocation(v)
				if err != nil {
					log.Fatalf("loader: can't add location %d: %s", v.ID, err)
				}
				atomic.AddInt32(&c.Locations, 1)
			}
		case key == strVisits:
			if stage != 2 {
				return
			}
			handler = func() {
				var v models.Visit
				err = decoder.Decode(&v)
				if err != nil {
					log.Fatalf("loader: bad visit: %s", err)
				}
				err = app.db.AddVisit(v)
				if err != nil {
					log.Fatalf("loader: can't add visit %d: %s", v.ID, err)
				}
				atomic.AddInt32(&c.Visits, 1)
			}
		default:
			log.Fatalf("loader: unknown section: %s", key)
		}

		log.Printf("loader: %s: loading %s", f.Name, key)

		for decoder.More() {
			handler()
		}
	}

}
