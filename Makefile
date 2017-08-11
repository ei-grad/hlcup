.PHONY: clean docker publish

docker: hlcup
	tar c hlcup | docker import - stor.highloadcup.ru/travels/raccoon_shooter

GENERATED = entities/location_cmap.go entities/locationmarks_cmap.go entities/types_ffjson.go entities/user_cmap.go entities/uservisits_cmap.go entities/visit_cmap.go

entities/location_cmap.go: entities/types.go
	rm -f $(GENERATED)
	go generate ./entities
	rm -rf ffjson-*

entities/locationmarks_cmap.go entities/types_ffjson.go entities/user_cmap.go entities/uservisits_cmap.go entities/visit_cmap.go: entities/location_cmap.go

hlcup: *.go */*.go $(GENERATED)
	CGO_ENABLED=0 go build -ldflags="-s -w"

run: docker
	docker run -it --rm -p 127.0.0.1:8000:80 -v $$PWD/data:/tmp/data stor.highloadcup.ru/travels/raccoon_shooter /hlcup

publish:
	docker push stor.highloadcup.ru/travels/raccoon_shooter

clean:
	go clean ./...
	rm -rf hlcup entities/ffjson-* $(GENERATED)
