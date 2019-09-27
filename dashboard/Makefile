.PHONY: build

build:
	( docker run -v $(shell pwd):/dashboard node:10.16.3-alpine /bin/sh -c "cd dashboard/client && yarn && yarn build")
