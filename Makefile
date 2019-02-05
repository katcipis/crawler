img=katcipis/crawler
run=docker run -ti -v `pwd`:/app $(img)

image:
	docker build . -t $(img)

shell: image
	$(run) sh
