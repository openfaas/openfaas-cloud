package test

import (
	"fmt"
	"strings"
	"testing"
)

func Test_IngressAuthNoTLS(t *testing.T) {
	parts := []string{
		"--set", "global.rootDomain=myfass.club",
	}

	want := buildIngressAuth("myfass.club", "auth.system", "openfaas-auth-ingress", "nginx", "600", "20", false)
	runYamlTest(parts, "./tmp/openfaas-cloud/templates/ingress/ingress-auth.yml", want, t)
}
func Test_IngressAuthWithTLS(t *testing.T) {
	parts := []string{
		"--set", "global.rootDomain=myfass.club",
		"--set", "tls.enabled=true",
	}

	want := buildIngressAuth("myfass.club", "auth.system", "openfaas-auth-ingress", "nginx", "600", "20", true)
	runYamlTest(parts, "./tmp/openfaas-cloud/templates/ingress/ingress-auth.yml", want, t)
}

func Test_IngressWildcardNoTLS(t *testing.T) {
	parts := []string{
		"--set", "global.rootDomain=myfass.club",
	}
	want := buildIngressAuth("myfass.club", "*", "openfaas-ingress", "nginx", "600", "20", false)
	runYamlTest(parts, "./tmp/openfaas-cloud/templates/ingress/ingress-wildcard.yml", want, t)
}
func Test_IngressWildcardWithTLS(t *testing.T) {
	parts := []string{
		"--set", "global.rootDomain=myfass.club",
		"--set", "tls.enabled=true",
	}

	want := buildIngressAuth("myfass.club", "*", "openfaas-ingress", "nginx", "600", "20", true)
	runYamlTest(parts, "./tmp/openfaas-cloud/templates/ingress/ingress-wildcard.yml", want, t)
}

func Test_IngressWildcardIngressClass(t *testing.T) {

	parts := []string{
		"--set", "global.rootDomain=myfass.club",
		"--set", "ingress.class=traefik",
	}

	want := buildIngressAuth("myfass.club", "*", "openfaas-ingress", "traefik", "600", "20", false)
	runYamlTest(parts, "./tmp/openfaas-cloud/templates/ingress/ingress-wildcard.yml", want, t)
}

func Test_IngressAuthIngressClass(t *testing.T) {

	parts := []string{
		"--set", "global.rootDomain=myfass.club",
		"--set", "ingress.class=traefik",
	}

	want := buildIngressAuth("myfass.club", "auth.system", "openfaas-auth-ingress", "traefik", "600", "20", false)
	runYamlTest(parts, "./tmp/openfaas-cloud/templates/ingress/ingress-auth.yml", want, t)
}

func Test_IngressAuthIngress_RPM(t *testing.T) {

	parts := []string{
		"--set", "global.rootDomain=myfass.club",
		"--set", "ingress.requestsPerMinute=200",
	}

	want := buildIngressAuth("myfass.club", "auth.system", "openfaas-auth-ingress", "nginx", "200", "20", false)
	runYamlTest(parts, "./tmp/openfaas-cloud/templates/ingress/ingress-auth.yml", want, t)
}

func Test_IngressAuthIngress_MaxCon(t *testing.T) {

	parts := []string{
		"--set", "global.rootDomain=myfass.club",
		"--set", "ingress.maxConnections=200",
	}

	want := buildIngressAuth("myfass.club", "auth.system", "openfaas-auth-ingress", "nginx", "600", "200", false)
	runYamlTest(parts, "./tmp/openfaas-cloud/templates/ingress/ingress-auth.yml", want, t)
}

func buildIngressAuth(hostDomain, prefix, name, ingressClass, rpm, maxConnections string, tls bool) YamlSpec {
	annotations := make(map[string]string)
	labels := make(map[string]string)
	backend := make(map[string]string)

	backend["serviceName"] = "edge-router"
	backend["servicePort"] = "8080"

	labels["app"] = "faas-netesd"

	annotations["kubernetes.io/ingress.class"] = ingressClass
	annotations["nginx.ingress.kubernetes.io/limit-connections"] = maxConnections
	annotations["nginx.ingress.kubernetes.io/limit-rpm"] = rpm
	if tls {
		annotations["cert-manager.io/issuer"] = "letsencrypt-staging"
	}

	spec := Spec{
		Rules: []SpecRules{{
			Host: fmt.Sprintf("%s.%s", prefix, hostDomain),
			Http: HttpPaths{
				Paths: []PathType{
					{
						Path:    "/",
						Backend: backend,
					},
				},
			},
		},
		},
	}
	if tls {
		spec.TLS = []TLSSpec{{
			Hosts:      []string{fmt.Sprintf("%s.%s", prefix, hostDomain)},
			SecretName: fmt.Sprintf("%s-%s-cert", strings.ReplaceAll(strings.ReplaceAll(prefix, ".", "-"), "*", "wildcard"), hostDomain),
		}}

	}

	return YamlSpec{
		ApiVersion: "extensions/v1beta1",
		Kind:       "Ingress",
		Metadata: MetadataItems{
			Name:        name,
			Namespace:   "openfaas",
			Annotations: annotations,
			Labels:      labels,
		},
		Spec: spec,
	}
}
