FROM golang:1.9
EXPOSE 80
WORKDIR /go/src/github.com/ei-grad/hlcup

RUN go get -u github.com/golang/dep/cmd/dep

ADD Gopkg.toml Gopkg.lock ./

RUN dep ensure -vendor-only

#CMD taskset -c 0 /go/bin/hlcup
CMD nice -20 /go/bin/hlcup
ENV RUN_TOP 1
#ENV GOMAXPROCS 1

ADD . .

RUN go get --ldflags="-X main.BuildDate=`date --iso=seconds`"
