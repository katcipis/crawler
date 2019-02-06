img=katcipis/crawler
run=docker run --rm -ti -v `pwd`:/app $(img)

image:
	docker build . -t $(img)

shell: image
	$(run) sh

check: image
	$(run) go test ./...
