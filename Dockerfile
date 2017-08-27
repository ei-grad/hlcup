FROM golang:1.9
EXPOSE 80
WORKDIR /go/src/github.com/ei-grad/hlcup

RUN curl https://glide.sh/get | sh

ADD glide.lock glide.yaml ./

RUN glide install

CMD nice -20 /go/bin/hlcup
ENV RUN_TOP 1
#ENV GOMAXPROCS 1
#CMD taskset -c 0 /go/bin/hlcup

ADD . .

RUN go get --ldflags="-X main.BuildDate=`date --iso=seconds`"
