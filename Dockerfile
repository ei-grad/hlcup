FROM golang:rc
CMD /go/bin/hlcup
EXPOSE 80
ENV RUN_TOP 1

WORKDIR /go/src/github.com/ei-grad/hlcup

RUN curl https://glide.sh/get | sh

ADD glide.lock glide.yaml ./

RUN glide install

ADD . .

RUN go get
