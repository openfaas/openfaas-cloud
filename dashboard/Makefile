.PHONY: build-dist build push

all:
	make build-dist build push

build-dist:
	( docker run -v $(shell pwd):/dashboard node:12-alpine /bin/sh -c "cd dashboard/client && yarn && yarn build")

build:
	faas-cli build

push:
	faas-cli push
