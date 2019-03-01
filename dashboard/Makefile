.PHONY: build

build:
	( docker run -v $(shell pwd):/dashboard node:10.20.0-alpine /bin/sh -c "cd dashboard/client && yarn && yarn build")
