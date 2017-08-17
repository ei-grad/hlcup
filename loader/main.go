package main

import "flag"

var baseURL = flag.String("url", "http://localhost", "base URL (for loader)")
var nWorkers = flag.Int("w", 8, "number of parallel requests while loading data")
var dataFileName = flag.String("data", "/tmp/data/data.zip", "data file name")

func main() {
	flag.Parse()
	LoadData(*baseURL, *dataFileName, *nWorkers)
}
