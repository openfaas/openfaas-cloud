.PHONY: build

build:
	( docker run -v $(shell pwd):/dashboard node:8-alpine /bin/sh -c "cd dashboard/client && yarn && yarn run build")
