.PHONY: build

build:
	( docker run -v $(shell pwd):/dashboard node:12-alpine /bin/sh -c "cd dashboard/client && yarn && yarn build")
