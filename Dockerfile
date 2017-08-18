FROM golang:rc
CMD /go/bin/hlcup
EXPOSE 80
ENV RUN_TOP 1

WORKDIR /go/src/github.com/ei-grad/hlcup

RUN curl https://glide.sh/get | sh && \
    mkdir -p /go/src/github.com/pquerna && cd /go/src/github.com/pquerna && \
    git clone --depth 1 -b hlcup https://github.com/ei-grad/ffjson && \
    cd ffjson && go install

ADD glide.lock glide.yaml ./

RUN glide install

ADD . .

RUN go get
