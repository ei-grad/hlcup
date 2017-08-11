docker: hlcup
	tar c hlcup | docker import - stor.highloadcup.ru/travels/raccoon_shooter

entities/types_ffjson.go: entities/types.go
	rm -f entities/types_ffjson.go
	go generate ./entities
	rm -rf entities/ffjson-*

maps/user_cmap.go: maps/maps.go
	go generate ./maps

maps/location_cmap.go maps/visit_cmap.go: maps/user_cmap.go

hlcup: *.go */*.go entities/types_ffjson.go maps/user_cmap.go maps/location_cmap.go maps/visit_cmap.go
	CGO_ENABLED=0 go build -ldflags="-s -w"

run: docker
	docker run -it --rm -p 127.0.0.1:8000:80 -v $$PWD/data:/tmp/data stor.highloadcup.ru/travels/raccoon_shooter /hlcup

publish:
	docker push stor.highloadcup.ru/travels/raccoon_shooter

clean:
	go clean
	rm -rf hlcup entities/ffjson-* entities/types_ffjson.go maps/*_cmap.go
