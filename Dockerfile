FROM golang:1.11.5-stretch

ENV GO111MODULE=on

WORKDIR /app

COPY go.mod ./go.mod
COPY go.sum ./go.sum

ENV GOLANG_CI_LINT_VERSION=v1.13.2

RUN cd /usr && \
    wget -O - -q https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | sh -s ${GOLANG_CI_LINT_VERSION}

RUN go mod download
