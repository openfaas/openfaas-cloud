TAG?=latest
.PHONY: all
all:
	cd contrib && ./ci.sh

.PHONY: ci-armhf-build
ci-armhf-build:
	./contrib/ci-arm.sh "build"

.PHONY: ci-armhf-push
ci-armhf-push:
	./contrib/ci-arm.sh "push"

.PHONY: ci-arm64-build
ci-arm64-build:
	./contrib/ci-arm.sh "build"

.PHONY: ci-arm64-push
ci-arm64-push:
	./contrib/ci-arm.sh "push"
