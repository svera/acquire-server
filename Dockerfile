FROM golang:1.7-alpine
# Git is needed for go get
RUN apk add --no-cache git
WORKDIR /go/src/github.com/svera/sackson-server
