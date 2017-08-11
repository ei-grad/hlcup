entities/types_ffjson.go: entities/types.go
	rm -rf entities/ffjson-* entities/*_ffjson.go
	go generate ./entities

hlcup: *.go */*.go entities/types_ffjson.go
	CGO_ENABLED=0 go build -ldflags="-s -w"

docker: hlcup
	tar c hlcup | docker import - stor.highloadcup.ru/travels/raccoon_shooter

run: docker
	docker run -it --rm -p 127.0.0.1:8000:80 -v $$PWD/data:/tmp/data stor.highloadcup.ru/travels/raccoon_shooter /hlcup

publish:
	docker push stor.highloadcup.ru/travels/raccoon_shooter

clean:
	rm hlcup
