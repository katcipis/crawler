img=katcipis/crawler
run=docker run --rm -ti -v `pwd`:/app $(img)
cov=coverage.out
covhtml=coverage.html

all: check build

image:
	docker build . -t $(img)

shell: image
	$(run) sh

build: image
	$(run) go build -o ./cmd/crawler/crawler ./cmd/crawler

check: image
	$(run) go test -timeout 60s -race -coverprofile=$(cov) ./...

check-integration: image
	$(run) go test -timeout 120s -race -coverprofile=$(cov) -tags=integration ./...

benchmark: image
	$(run) go test ./... -run=NONE -timeout 1h -bench=.

coverage: check
	$(run) go tool cover -html=$(cov) -o=$(covhtml)
	xdg-open coverage.html

static-analysis: image
	$(run) golangci-lint run ./...

graph: build
	$(run) ./tools/graph $(url)
	xdg-open sitemap.dot.png
