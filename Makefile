TAG?=latest

all:
	cd contrib && ./ci.sh

charts:
	cd chart && helm package openfaas-cloud/
	mv chart/*.tgz docs/
	helm repo index docs --url https://openfaas-cloud.github.io/openfaas-cloud/ --merge ./docs/index.yaml

test-chart:
	cd chart/test && go test -v -mod=vendor ./...