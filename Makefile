.PHONY: clean run generated fixlinter watch publish

IMAGE = stor.highloadcup.ru/travels/raccoon_shooter

all: fixlinter hlcup

docker: $(SOURCES) $(wildcard Dockerfile*)
	rm -f Dockerfile
	ln -s Dockerfile.$(DOCKERFILE) Dockerfile
	docker build -t $(IMAGE):$(DOCKERFILE) .
	docker tag $(IMAGE):$(DOCKERFILE) $(IMAGE):latest
	touch docker

fixlinter: generated
	# "Running 'go get' to fix linters analysis"
	go clean github.com/ei-grad/hlcup/...
	go get github.com/ei-grad/hlcup/...

DB = array

ifeq ($(DB), cmap)
GENERATED = \
	models/entities_ffjson.go \
	models/indexes_ffjson.go \
	models/location_cmap.go \
	models/locationmarks_cmap.go \
	models/user_cmap.go \
	models/uservisits_cmap.go \
	models/visit_cmap.go
else
TAGS = -tags db_use_array
GENERATED = \
	models/entities_ffjson.go \
	models/indexes_ffjson.go
endif

$(GENERATED): models/entities.go models/indexes.go
	rm -f $(GENERATED)
	go generate ./models
	rm -rf models/ffjson-*

generated: $(GENERATED)

DATE = $(shell LANG=C date --iso=seconds)
APP_VERSION = $(shell git describe --tags)
LDFLAGS = '-s -w -X main.appVersion=$(APP_VERSION)/DB=$(DB) -X main.appBuildDate=$(DATE)'
#LDFLAGS = '-X main.appVersion=$(APP_VERSION)/DB=$(DB) -X main.appBuildDate=$(DATE)'
SOURCES = $(wildcard *.go */*.go)

hlcup: $(SOURCES) $(GENERATED)
	CGO_ENABLED=0 go build $(TAGS) -ldflags=$(LDFLAGS)

DATA = full

race: $(SOURCES) $(GENERATED)
	go run -race $(TAGS) -ldflags=$(LDFLAGS) $(wildcard *.go) -b :8000 -data $(DATA)/data.zip $(ARGS)

run: docker
	docker run -it --rm --net=host -v `realpath $(DATA)`:/tmp/data $(IMAGE) hlcup $(ARGS)

publish: docker
	docker push $(IMAGE)

clean:
	go clean ./... github.com/ei-grad/hlcup/...
	rm -rf hlcup docker models/ffjson-inception* models/*_ffjson_expose.go $(GENERATED)

watch: $(SOURCES) $(GENERATED)
	iwatch "go build $(TAGS) -ldflags=$(LDFLAGS) -o hlcup-watch && ./hlcup-watch -b :8000 -data $(DATA)/data.zip $(ARGS)"
