module github.com/openfaas/openfaas-cloud/pipeline-log

go 1.13

replace (
	github.com/openfaas/openfaas-cloud/pipeline-log => /home/heyal/go/src/github.com/openfaas/openfaas-cloud/pipeline-log
	github.com/openfaas/openfaas-cloud/sdk => /home/heyal/go/src/github.com/openfaas/openfaas-cloud/sdk
)

require (
	github.com/minio/minio-go/v6 v6.0.57
	github.com/openfaas/faas-provider v0.0.0-20191011092439-98c25c3919da // indirect
	github.com/openfaas/openfaas-cloud/sdk v0.0.0-00010101000000-000000000000
	golang.org/x/crypto v0.0.0-20200302210943-78000ba7a073 // indirect
	golang.org/x/net v0.0.0-20200301022130-244492dfa37a // indirect
	golang.org/x/sys v0.0.0-20200302150141-5c8b2ff67527 // indirect
	golang.org/x/text v0.3.2 // indirect
	gopkg.in/ini.v1 v1.52.0 // indirect
)
