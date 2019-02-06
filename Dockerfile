FROM golang:1.11.5-alpine3.8

RUN apk update && \
    apk add git

WORKDIR /app

COPY go.mod ./go.mod
COPY go.sum ./go.sum

RUN go mod download
