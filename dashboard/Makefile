.PHONY: build

build:
	( docker run -v $(shell pwd):/dashboard node:10.12.0-alpine /bin/sh -c "cd dashboard/client && yarn && yarn build")
