FROM golang:1.8-alpine
# Git is needed for go get
RUN apk add --no-cache git gcc libc-dev
COPY . /go/src/github.com/svera/sackson-server
WORKDIR /go/src/github.com/svera/sackson-server
RUN mkdir /usr/lib/sackson-server
COPY acquire.so /usr/lib/sackson-server
#RUN ./download_drivers.sh
RUN go get github.com/kardianos/govendor
RUN go get github.com/pilu/fresh
