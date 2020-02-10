package test

import (
	"fmt"
	"io/ioutil"
	"testing"
)

func Test_IssuerStaging_NotEnabled(t *testing.T) {
	parts := []string{
		"--set", "tls.enabled=false",
	}

	filename := "./tmp/no-tls-staging/openfaas-cloud/templates/tls/issuer-staging.yml"
	_, _ = helmRunnerToLocation("./tmp/no-tls-staging", parts...)

	_, err := ioutil.ReadFile(filename)
	if err == nil {
		t.Errorf("expect error when reading yaml, as we disabled tls. got: nil")
	}
}

func Test_IssuerStaging_Route53(t *testing.T) {
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

	want := makeIssuer(tlsSettings, "openfaas", "letsencrypt-staging")
	runYamlTest(parts, "./tmp/openfaas-cloud/templates/tls/issuer-staging.yml", want, t)
}

func Test_IssuerStaging_Route53_NonDefaultCoreNamespace(t *testing.T) {
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

	want := makeIssuer(tlsSettings, "some-core-ns", "letsencrypt-staging")
	runYamlTest(parts, "./tmp/openfaas-cloud/templates/tls/issuer-staging.yml", want, t)
}

func Test_IssuerStaging_Route53_AmbientCredentials(t *testing.T) {
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

	want := makeIssuer(tlsSettings, "openfaas", "letsencrypt-staging")
	runYamlTest(parts, "./tmp/openfaas-cloud/templates/tls/issuer-staging.yml", want, t)
}

func Test_IssuerStaging_DigitalOcean(t *testing.T) {
	tlsSettings := TLSSettings{
		dnsService: "digitalocean",
		email:      "example@example.com",
	}
	parts := []string{
		"--set", "tls.enabled=true",
		"--set", fmt.Sprintf("tls.email=%s", tlsSettings.email),
		"--set", fmt.Sprintf("tls.dnsService=%s", tlsSettings.dnsService),
	}

	want := makeIssuer(tlsSettings, "openfaas", "letsencrypt-staging")
	runYamlTest(parts, "./tmp/openfaas-cloud/templates/tls/issuer-staging.yml", want, t)

}

func Test_IssuerStaging_CloudDns(t *testing.T) {
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

	want := makeIssuer(tlsSettings, "openfaas", "letsencrypt-staging")
	runYamlTest(parts, "./tmp/openfaas-cloud/templates/tls/issuer-staging.yml", want, t)

}

func Test_IssuerStaging_Cloudflare(t *testing.T) {
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

	want := makeIssuer(tlsSettings, "openfaas", "letsencrypt-staging")
	runYamlTest(parts, "./tmp/openfaas-cloud/templates/tls/issuer-staging.yml", want, t)

}
