FROM golang:1.9-alpine
# Git is needed for go get
RUN apk add --no-cache git gcc libc-dev curl ca-certificates openssl
COPY . /go/src/github.com/svera/sackson-server
WORKDIR /usr/lib/sackson-server
RUN wget -qO acquire.so $(\
    curl -s https://api.github.com/repos/svera/acquire-sackson-driver/releases/latest \
  | grep "browser_download_url.*so" \
  | cut -d '"' -f 4 )
COPY ./drivers/acquire.so /usr/lib/sackson-server/acquire.so
WORKDIR /go/src/github.com/svera/sackson-server
RUN go get github.com/kardianos/govendor
RUN go get github.com/pilu/fresh
