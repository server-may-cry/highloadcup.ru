FROM golang:1.8-alpine AS build-env
ADD . /go/src/github.com/server-may-cry/highloadcup.ru
WORKDIR /go/src/github.com/server-may-cry/highloadcup.ru
RUN apk update && apk upgrade && \
    apk add --no-cache git openssl make && \
    go get -u github.com/golang/dep/cmd/dep && \
    dep ensure && \
    CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo .

FROM scratch
COPY --from=build-env /go/src/github.com/server-may-cry/highloadcup.ru/highloadcup.ru /
CMD ["/server"]
