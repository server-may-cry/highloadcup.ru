FROM golang:1.9-alpine AS build-env
RUN apk update && apk upgrade && \
    apk add --no-cache git openssl && \
    go get -u github.com/golang/dep/cmd/dep

ADD . /go/src/github.com/server-may-cry/highloadcup.ru
WORKDIR /go/src/github.com/server-may-cry/highloadcup.ru

RUN dep ensure && \
    GOGC=off CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo .

FROM scratch
COPY --from=build-env /go/src/github.com/server-may-cry/highloadcup.ru/highloadcup.ru /server
CMD ["/server"]
