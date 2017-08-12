.PHONY: clean run generated fixlinter publish

all: docker fixlinter

docker: hlcup Dockerfile
	docker build -t stor.highloadcup.ru/travels/raccoon_shooter .
	touch docker

fixlinter: generated
	# "Running 'go get' to fix linters analysis"
	go get ./...

GENERATED = \
	models/entities_ffjson.go \
	models/location_cmap.go \
	models/locationmarks_cmap.go \
	models/user_cmap.go \
	models/uservisits_cmap.go \
	models/visit_cmap.go

models/entities_ffjson.go: models/entities.go models/indexes.go
	rm -f $(GENERATED)
	go generate ./models
	rm -rf ffjson-*

models/location_cmap.go models/locationmarks_cmap.go models/user_cmap.go models/uservisits_cmap.go models/visit_cmap.go: models/entities_ffjson.go 

generated: $(GENERATED)

hlcup: *.go */*.go $(GENERATED)
	CGO_ENABLED=0 go build -ldflags="-s -w"

run: docker
	docker run -it --rm -p 127.0.0.1:80:80 -v $$PWD/data:/tmp/data stor.highloadcup.ru/travels/raccoon_shooter ./hlcup -v

publish:
	docker push stor.highloadcup.ru/travels/raccoon_shooter

clean:
	go clean ./... github.com/ei-grad/hlcup/...
	rm -rf hlcup models/ffjson-* $(GENERATED)
