TAG?=latest

build:
	docker build --build-arg http_proxy="${http_proxy}" --build-arg https_proxy="${https_proxy}" -t openfaas/cloud-auth:$(TAG) .

push:
	docker push openfaas/cloud-auth:$(TAG)

ci-armhf-build:
	docker build --build-arg http_proxy="${http_proxy}" --build-arg https_proxy="${https_proxy}" -t openfaas/cloud-auth:$(TAG)-armhf .

ci-armhf-push:
	docker push openfaas/cloud-auth:$(TAG)-armhf

ci-arm64-build:
	docker build --build-arg http_proxy="${http_proxy}" --build-arg https_proxy="${https_proxy}" -t openfaas/cloud-auth:$(TAG)-arm64 .

ci-arm64-push:
	docker push openfaas/cloud-auth:$(TAG)-arm64
