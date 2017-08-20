FROM golang:1.8
CMD /go/bin/hlcup
EXPOSE 80
ENV RUN_TOP 1

WORKDIR /go/src/github.com/ei-grad/hlcup

RUN curl https://glide.sh/get | sh

ADD glide.lock glide.yaml ./

RUN glide install

ADD . .

RUN go get --ldflags="-X main.BuildDate=`date --iso=seconds`"

ENV GOMAXPROCS 1
CMD taskset -c 0 /go/bin/hlcup
