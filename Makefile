.PHONY: clean run generated fixlinter watch publish

IMAGE = stor.highloadcup.ru/travels/raccoon_shooter

all: fixlinter hlcup

fixlinter: generated
	# "Running 'go get' to fix linters analysis"
	go clean github.com/ei-grad/hlcup/...
	go get github.com/ei-grad/hlcup/...

DB = array

GENERATED = \
	models/entities_easyjson.go \
	models/indexes_easyjson.go

$(GENERATED): models/entities.go models/indexes.go
	go generate ./models

generated: $(GENERATED)

DATE = $(shell LANG=C date --iso=seconds)
LDFLAGS = '-s -w -X main.appBuildDate=$(DATE)'
SOURCES = $(wildcard *.go */*.go)

hlcup: $(SOURCES) $(GENERATED)
	CGO_ENABLED=0 go build $(TAGS) -ldflags=$(LDFLAGS)

docker: $(SOURCES) Dockerfile
	docker build -t $(IMAGE) .
	touch docker

DATA = full

race: $(SOURCES) $(GENERATED)
	go run -race $(TAGS) -ldflags=$(LDFLAGS) $(wildcard *.go) -b :8000 -data $(DATA)/data.zip $(ARGS)

run: docker
	docker run -it --rm --net=host -v `realpath $(DATA)`:/tmp/data $(IMAGE)

publish: docker
	docker push $(IMAGE)

clean:
	go clean ./... github.com/ei-grad/hlcup/...
	rm -rf hlcup docker $(GENERATED)

watch: $(SOURCES) $(GENERATED)
	iwatch "go build $(TAGS) -ldflags=$(LDFLAGS) -o hlcup-watch && ./hlcup-watch -b :8000 -data $(DATA)/data.zip $(ARGS)"
