FROM golang:1.11.5-stretch

ENV GO111MODULE=on

ENV GOLANG_CI_LINT_VERSION=v1.13.2

RUN apt-get update && \
    apt-get install -y graphviz

RUN cd /usr && \
    wget -O - -q https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | sh -s ${GOLANG_CI_LINT_VERSION}

WORKDIR /app

COPY go.mod ./go.mod
COPY go.sum ./go.sum

RUN go mod download
