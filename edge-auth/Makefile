TAG?=latest
NAMESPACE?=openfaas
build:
	docker build --build-arg http_proxy="${http_proxy}" --build-arg https_proxy="${https_proxy}" -t $(NAMESPACE)/edge-auth:$(TAG) .

push:
	docker push $(NAMESPACE)/edge-auth:$(TAG)
