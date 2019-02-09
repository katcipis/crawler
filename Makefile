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

coverage: check
	$(run) go tool cover -html=$(cov) -o=$(covhtml)
	xdg-open coverage.html

static-analysis: image
	$(run) golangci-lint run ./...

graph: build
	$(run) ./cmd/crawler/crawler -url $(url) -format graphviz 1> sitemap.dot 2> errors.log && \
	$(run) dot sitemap.dot -Tpng > sitemap.png && \
	xdg-open sitemap.png
