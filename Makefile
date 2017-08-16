.PHONY: clean run generated fixlinter watch publish

IMAGE = stor.highloadcup.ru/travels/raccoon_shooter

all: fixlinter hlcup

docker: hlcup Dockerfile
	docker build -t $(IMAGE) .
	touch docker

fixlinter: generated
	# "Running 'go get' to fix linters analysis"
	go clean github.com/ei-grad/hlcup/...
	go get github.com/ei-grad/hlcup/...

GENERATED = \
	models/entities_ffjson.go \
	models/indexes_ffjson.go \
	models/location_cmap.go \
	models/locationmarks_cmap.go \
	models/user_cmap.go \
	models/uservisits_cmap.go \
	models/visit_cmap.go

$(GENERATED): models/entities.go models/indexes.go
	rm -f $(GENERATED)
	go generate ./models
	rm -rf models/ffjson-*

generated: $(GENERATED)

DATE = $(shell LANG=C date --iso=seconds)
APP_VERSION = $(shell git describe --tags)
LDFLAGS = '-s -w -X main.appVersion=$(APP_VERSION) -X main.appBuildDate=$(DATE)'
SOURCES = $(wildcard *.go */*.go)

hlcup: $(SOURCES) $(GENERATED)
	CGO_ENABLED=0 go build -ldflags=$(LDFLAGS)

DATADIR = train

run: docker
	docker run -it --rm --net=host -v `realpath $(DATADIR)`:/tmp/data $(IMAGE) ./hlcup $(ARGS)

publish: docker
	docker push $(IMAGE)

clean:
	go clean ./... github.com/ei-grad/hlcup/...
	rm -rf hlcup models/ffjson-inception* models/*_ffjson_expose.go $(GENERATED)

watch: $(GENERATED)
	iwatch "go build -race -o debug && ./debug -b :8000 -url http://127.0.0.1:8000 -data data/data.zip -v"
