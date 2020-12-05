module github.com/openfaas/openfaas-cloud

go 1.13

require (
	code.cloudfoundry.org/bytefmt v0.0.0-20200131002437-cf55d5288a48
	github.com/alexellis/derek v0.0.0-20201203223145-52084a5968ea
	github.com/alexellis/hmac v0.0.0-20180624211220-5c52ab81c0de
	github.com/aws/aws-sdk-go v1.36.2
	github.com/bitnami-labs/sealed-secrets v0.13.1
	github.com/go-ini/ini v1.62.0 // indirect
	github.com/google/go-github v17.0.0+incompatible
	github.com/minio/minio-go v6.0.14+incompatible
	github.com/openfaas/faas v0.0.0-20200422142642-18f6c720b50d
	github.com/openfaas/faas-cli v0.0.0-20201203202533-429edae5124b
	github.com/openfaas/openfaas-cloud/metrics v0.0.0-20201201105924-2f2413a8b8ab // indirect
	github.com/openfaas/openfaas-cloud/sdk v0.0.0-20201205095205-64e6798b27b1
	gopkg.in/yaml.v2 v2.4.0
	k8s.io/apimachinery v0.19.4
	k8s.io/client-go v11.0.0+incompatible
)
