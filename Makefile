img=katcipis/crawler
run=docker run --rm -ti -v `pwd`:/app $(img)
cov=coverage.out
covhtml=coverage.html

image:
	docker build . -t $(img)

shell: image
	$(run) sh

check: image
	$(run) go test -timeout 60s -race -coverprofile=$(cov) ./...

coverage: check
	$(run) go tool cover -html=$(cov) -o=$(covhtml)
	xdg-open coverage.html

static-analysis: image
	$(run) golangci-lint run ./...
