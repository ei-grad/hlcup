hlcup: *.go */*.go
	CGO_ENABLED=0 go build -ldflags="-s -w"

docker: hlcup
	tar c hlcup | docker import - stor.highloadcup.ru/travels/raccoon_shooter

run: docker
	docker run -it --rm -p 127.0.0.1:8000:80 stor.highloadcup.ru/travels/raccoon_shooter /hlcup

publish:
	docker push stor.highloadcup.ru/travels/raccoon_shooter

clean:
	rm hlcup
