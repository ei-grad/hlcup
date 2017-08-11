package main

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/valyala/fasthttp"

	"github.com/ei-grad/hlcup/entities"
)

func loadData() {

	// Wait for a server to start
	time.Sleep(5)

	fileName := os.Getenv("DATA")
	if fileName == "" {
		fileName = "/tmp/data/data.zip"
	}

	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost"
	}

	// Open a zip archive for reading.
	r, err := zip.OpenReader(fileName)
	if err != nil {
		log.Fatal(err)
	}
	defer r.Close()

	// Iterate through the files in the archive,
	// printing some of their contents.
	for _, f := range r.File {

		log.Printf("Loading %s...", f.Name)
		rc, err := f.Open()
		if err != nil {
			log.Fatal(err)
		}

		decoder := json.NewDecoder(rc)

		// read left_bracket token
		token, err := decoder.Token()
		if err != nil {
			log.Fatalf("Bad start token in %s!", f.Name)
		}
		if _, ok := token.(json.Delim); !ok {
			log.Fatalf("Bad start token in %s!", f.Name)
		}

		// read key
		token, err = decoder.Token()
		if err != nil {
			log.Fatalf("Bad second token in %s!", f.Name)
		}
		key, ok := token.(string)
		if !ok {
			log.Fatalf("Second token in %s is not string!", f.Name)
		}

		// read left_brace token
		token, err = decoder.Token()
		if err != nil {
			log.Fatalf("Bad start token in %s!", f.Name)
		}
		if _, ok := token.(json.Delim); !ok {
			log.Fatalf("Bad start token in %s!", f.Name)
		}

		switch key {
		case "users":
			for decoder.More() {
				var v entities.User
				err := decoder.Decode(&v)
				if err != nil {
					log.Fatalf("Bad JSON: %s", err)
				}
				body, err := v.MarshalJSON()
				if err != nil {
					log.Fatalf("Can't encode %+v back: %s", v, err)
				}
				sendPost(fmt.Sprintf("%s/users/new", baseURL), body)
			}
		case "locations":
			for decoder.More() {
				var v entities.Location
				err := decoder.Decode(&v)
				if err != nil {
					log.Fatalf("Bad JSON: %s", err)
				}
				body, err := v.MarshalJSON()
				if err != nil {
					log.Fatalf("Can't encode %+v back: %s", v, err)
				}
				sendPost(fmt.Sprintf("%s/locations/new", baseURL), body)
			}
		case "visits":
			for decoder.More() {
				var v entities.Visit
				err := decoder.Decode(&v)
				if err != nil {
					log.Fatalf("Bad JSON: %s", err)
				}
				body, err := v.MarshalJSON()
				if err != nil {
					log.Fatalf("Can't encode %+v back: %s", v, err)
				}
				sendPost(fmt.Sprintf("%s/visits/new", baseURL), body)
			}
		}

		rc.Close()

		log.Printf("Loaded %s.", f.Name)
	}

}

func sendPost(url string, body []byte) {

	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	req.Header.SetMethod(strPost)
	req.SetRequestURI(url)
	req.Header.SetContentType(strApplicationJSON)
	req.SetBody(body)

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	err := fasthttp.Do(req, resp)
	if err != nil {
		panic(err)
	}

	if resp.StatusCode() != 200 {
		log.Fatal(resp)
	}

}
