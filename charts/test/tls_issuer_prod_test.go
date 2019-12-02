package test

import (
	"fmt"
	"io/ioutil"
	"testing"
)

func Test_IssuerProd_NotEnabled(t *testing.T) {
	parts := []string{
		"--set", "tls.enabled=false",
	}

	filename := "./tmp/no-tls-prod/openfaas-cloud/templates/tls/issuer-prod.yml"
	_, _ = helmRunnerToLocation("./tmp/no-tls-prod", parts...)

	_ , err := ioutil.ReadFile(filename)
	if err == nil {
		t.Errorf("expect error when reading yaml, as we disabled tls. got: nil")
	}
}

func Test_IssuerProd_Route53 (t *testing.T) {
	tlsSettings := TLSSettings{
		dnsService:  "route53",
		region:      "eu-west-1",
		email:       "example@example.com",
		accessKeyID: "ABC123",
	}

	parts := []string{
		"--set", "tls.enabled=true",
		"--set", fmt.Sprintf("tls.email=%s",tlsSettings.email),
		"--set", fmt.Sprintf("tls.dnsService=%s",tlsSettings.dnsService),
		"--set", fmt.Sprintf("tls.route53.region=%s", tlsSettings.region),
		"--set", fmt.Sprintf("tls.route53.accessKeyID=%s", tlsSettings.accessKeyID),
	}

	want := makeIssuerProd(tlsSettings, "openfaas")
	runYamlTest(parts, "./tmp/openfaas-cloud/templates/tls/issuer-prod.yml", want, t)
}

func makeIssuerProd(tlsSettings TLSSettings, coreNamespace string) YamlSpec {

	return YamlSpec{
		ApiVersion: "cert-manager.io/v1alpha2",
		Kind:       "ClusterIssuer",
		Metadata: MetadataItems{
			Name:        "letsencrypt-prod",
			Namespace:   coreNamespace,
			Annotations: map[string]string{"dnsProvider": tlsSettings.dnsService},
		},
		Spec: Spec{
			ACME: ACME{
				Email:               tlsSettings.email,
				Server:              "https://acme-v02.api.letsencrypt.org/directory",
				PrivateKeySecretRef: map[string]string{"name": "letsencrypt-prod"},
				Solvers:             []Solver{{
					DNSSolver: DNSSolver{
						Route53:Route53Solver{
							Region:                   tlsSettings.region,
							AccessKeyID:              tlsSettings.accessKeyID,
							SecretAccessKeySecretRef: map[string]string{
								"name": "route53-credentials-secret",
								"key": "secret-access-key",
							},
						},
					}},
				},
			},
		},
	}
}


type TLSSettings struct {
	dnsService string
	region string
	email string
	accessKeyID string
}