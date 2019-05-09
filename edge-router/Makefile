TAG?=latest

build:
	docker build --build-arg http_proxy="${http_proxy}" --build-arg https_proxy="${https_proxy}" -t openfaas/edge-router:$(TAG) .

push:
	docker push openfaas/edge-router:$(TAG)
