FROM golang:1.11.5-alpine3.8

RUN apk update && \
    apk add git gcc musl-dev

ENV GO111MODULE=on

WORKDIR /app

COPY go.mod ./go.mod
COPY go.sum ./go.sum

RUN go get -u honnef.co/go/tools/cmd/staticcheck@2019.1

RUN go mod download
