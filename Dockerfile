FROM golang:1.9-alpine
# Git is needed for go get
RUN apk add --no-cache git gcc libc-dev
COPY . /go/src/github.com/svera/sackson-server
COPY ./bridges/acquire.so /usr/lib/sackson/acquire.so
WORKDIR /go/src/github.com/svera/sackson-server
RUN go get github.com/kardianos/govendor
RUN go get github.com/pilu/fresh
