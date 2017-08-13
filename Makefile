.PHONY: clean run generated fixlinter watch publish

all: fixlinter hlcup

docker: hlcup Dockerfile
	docker build -t stor.highloadcup.ru/travels/raccoon_shooter .
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

hlcup: *.go */*.go $(GENERATED)
	CGO_ENABLED=0 go build -ldflags="-s -w"

run: docker
	docker run -it --rm --net=host -v $$PWD/data:/tmp/data stor.highloadcup.ru/travels/raccoon_shooter ./hlcup $(ARGS)

publish: docker
	docker push stor.highloadcup.ru/travels/raccoon_shooter

clean:
	go clean ./... github.com/ei-grad/hlcup/...
	rm -rf hlcup models/ffjson-* $(GENERATED)

watch: $(GENERATED)
	iwatch "go build -race -o debug && ./debug -b :8000 -url http://127.0.0.1:8000 -data data/data.zip -v"
