module github.com/openfaas/openfaas-cloud

go 1.13

require (
	code.cloudfoundry.org/bytefmt v0.0.0-20200131002437-cf55d5288a48
	github.com/alexellis/derek v0.0.0-20201101113259-57106cd1c26b
	github.com/alexellis/hmac v0.0.0-20180624211220-5c52ab81c0de
	github.com/aws/aws-sdk-go v1.35.34
	github.com/bitnami-labs/sealed-secrets v0.13.1
	github.com/go-ini/ini v1.62.0 // indirect
	github.com/google/go-github v17.0.0+incompatible
	github.com/minio/minio-go v6.0.14+incompatible
	github.com/openfaas/faas v0.0.0-20201117113642-9fccc1c84d6a
	github.com/openfaas/faas-cli v0.0.0-20201119120128-c9d284d0c5bd
	github.com/openfaas/openfaas-cloud/edge-auth v0.0.0-20201123101040-becb6362be5e // indirect
	github.com/openfaas/openfaas-cloud/sdk v0.0.0-20201123101040-becb6362be5e
	gopkg.in/yaml.v2 v2.3.0
	k8s.io/apimachinery v0.19.4
	k8s.io/client-go v11.0.0+incompatible
)
