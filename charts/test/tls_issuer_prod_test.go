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
	_, _ = helmRunnerToLocation("./tmp", parts...)

	_, err := ioutil.ReadFile(filename)
	if err == nil {
		t.Errorf("expect error when reading yaml, as we disabled tls. got: nil")
	}
}

func Test_IssuerProd_Route53(t *testing.T) {
	tlsSettings := TLSSettings{
		dnsService:  "route53",
		region:      "eu-west-1",
		email:       "example@example.com",
		accessKeyID: "ABC123",
	}

	parts := []string{
		"--set", "tls.enabled=true",
		"--set", fmt.Sprintf("tls.email=%s", tlsSettings.email),
		"--set", fmt.Sprintf("tls.dnsService=%s", tlsSettings.dnsService),
		"--set", fmt.Sprintf("tls.route53.region=%s", tlsSettings.region),
		"--set", fmt.Sprintf("tls.route53.accessKeyID=%s", tlsSettings.accessKeyID),
	}

	want := makeIssuer(tlsSettings, "openfaas", "letsencrypt-prod")
	runYamlTest(parts, "./tmp/openfaas-cloud/templates/tls/issuer-prod.yml", want, t)
}

func Test_IssuerProd_Route53_NonDefaultCoreNamespace(t *testing.T) {
	tlsSettings := TLSSettings{
		dnsService:  "route53",
		region:      "eu-west-1",
		email:       "example@example.com",
		accessKeyID: "ABC123",
	}

	parts := []string{
		"--set", "tls.enabled=true",
		"--set", "global.coreNamespace=some-core-ns",
		"--set", fmt.Sprintf("tls.email=%s", tlsSettings.email),
		"--set", fmt.Sprintf("tls.dnsService=%s", tlsSettings.dnsService),
		"--set", fmt.Sprintf("tls.route53.region=%s", tlsSettings.region),
		"--set", fmt.Sprintf("tls.route53.accessKeyID=%s", tlsSettings.accessKeyID),
	}

	want := makeIssuer(tlsSettings, "some-core-ns", "letsencrypt-prod")
	runYamlTest(parts, "./tmp/openfaas-cloud/templates/tls/issuer-prod.yml", want, t)
}

func Test_IssuerProd_Route53_AmbientCredentials(t *testing.T) {
	tlsSettings := TLSSettings{
		dnsService:                "route53",
		region:                    "eu-west-1",
		email:                     "example@example.com",
		route53AmbientCredentials: true,
	}

	parts := []string{
		"--set", "tls.enabled=true",
		"--set", fmt.Sprintf("tls.email=%s", tlsSettings.email),
		"--set", fmt.Sprintf("tls.dnsService=%s", tlsSettings.dnsService),
		"--set", fmt.Sprintf("tls.route53.region=%s", tlsSettings.region),
		"--set", "tls.route53.ambientCredentials=true",
	}

	want := makeIssuer(tlsSettings, "openfaas", "letsencrypt-prod")
	runYamlTest(parts, "./tmp/openfaas-cloud/templates/tls/issuer-prod.yml", want, t)
}

func Test_IssuerProd_DigitalOcean(t *testing.T) {
	tlsSettings := TLSSettings{
		dnsService: "digitalocean",
		email:      "example@example.com",
	}
	parts := []string{
		"--set", "tls.enabled=true",
		"--set", fmt.Sprintf("tls.email=%s", tlsSettings.email),
		"--set", fmt.Sprintf("tls.dnsService=%s", tlsSettings.dnsService),
	}

	want := makeIssuer(tlsSettings, "openfaas", "letsencrypt-prod")
	runYamlTest(parts, "./tmp/openfaas-cloud/templates/tls/issuer-prod.yml", want, t)

}

func Test_IssuerProd_CloudDns(t *testing.T) {
	tlsSettings := TLSSettings{
		dnsService:        "clouddns",
		email:             "example@example.com",
		cloudDnsProjectId: "ABCDEFG123",
	}
	parts := []string{
		"--set", "tls.enabled=true",
		"--set", fmt.Sprintf("tls.email=%s", tlsSettings.email),
		"--set", fmt.Sprintf("tls.dnsService=%s", tlsSettings.dnsService),
		"--set", fmt.Sprintf("tls.clouddns.projectID=%s", tlsSettings.cloudDnsProjectId),
	}

	want := makeIssuer(tlsSettings, "openfaas", "letsencrypt-prod")
	runYamlTest(parts, "./tmp/openfaas-cloud/templates/tls/issuer-prod.yml", want, t)

}

func Test_IssuerProd_Cloudflare(t *testing.T) {
	tlsSettings := TLSSettings{
		dnsService:      "cloudflare",
		email:           "example@example.com",
		cloudflareEmail: "someone@example.com",
	}
	parts := []string{
		"--set", "tls.enabled=true",
		"--set", fmt.Sprintf("tls.email=%s", tlsSettings.email),
		"--set", fmt.Sprintf("tls.dnsService=%s", tlsSettings.dnsService),
		"--set", fmt.Sprintf("tls.cloudflare.email=%s", tlsSettings.cloudflareEmail),
	}

	want := makeIssuer(tlsSettings, "openfaas", "letsencrypt-prod")
	runYamlTest(parts, "./tmp/openfaas-cloud/templates/tls/issuer-prod.yml", want, t)

}

func makeIssuer(tlsSettings TLSSettings, coreNamespace, issuerName string) YamlSpec {
	serverApi := "https://acme-v02.api.letsencrypt.org/directory"
	if issuerName == "letsencrypt-staging" {
		serverApi = "https://acme-staging-v02.api.letsencrypt.org/directory"
	}

	solver := makeSolver(tlsSettings)

	return YamlSpec{
		ApiVersion: "cert-manager.io/v1alpha2",
		Kind:       "ClusterIssuer",
		Metadata: MetadataItems{
			Name:        issuerName,
			Namespace:   coreNamespace,
			Annotations: map[string]string{"dnsProvider": tlsSettings.dnsService},
		},
		Spec: Spec{
			ACME: ACME{
				Email:               tlsSettings.email,
				Server:              serverApi,
				PrivateKeySecretRef: map[string]string{"name": issuerName},
				Solvers: []Solver{{
					DNSSolver: solver,
				}},
			},
		},
	}
}

func makeSolver(settings TLSSettings) DNSSolver {
	if settings.dnsService == "route53" {
		if settings.route53AmbientCredentials {
			return DNSSolver{
				Route53: Route53Solver{
					Region: settings.region,
				},
			}
		}
		return DNSSolver{
			Route53: Route53Solver{
				Region:      settings.region,
				AccessKeyID: settings.accessKeyID,
				SecretAccessKeySecretRef: map[string]string{
					"name": "route53-credentials-secret",
					"key":  "secret-access-key",
				},
			},
		}
	} else if settings.dnsService == "digitalocean" {
		return DNSSolver{
			DigitalOcean: DigitalOceanSolver{
				TokenSecretRef: map[string]string{
					"name": "digitalocean-dns",
					"key":  "access-token",
				},
			},
		}
	} else if settings.dnsService == "cloudflare" {
		return DNSSolver{
			Cloudflare: CloudflareSolver{
				Email: settings.cloudflareEmail,
				APIKeySecretRef: map[string]string{
					"name": "cloudflare-api-key-secret",
					"key":  "api-key",
				},
			},
		}

	} else if settings.dnsService == "clouddns" {
		return DNSSolver{
			CloudDNS: CloudDNSSolver{
				Project: settings.cloudDnsProjectId,
				ServiceAccountSecretRef: map[string]string{
					"name": "clouddns-service-account",
					"key":  "service-account.json",
				},
			},
		}
	}

	return DNSSolver{}
}

type TLSSettings struct {
	dnsService                string
	region                    string
	email                     string
	accessKeyID               string
	route53AmbientCredentials bool
	cloudflareEmail           string
	cloudDnsProjectId         string
}
