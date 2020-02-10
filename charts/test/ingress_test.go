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

	want := buildIngressAuth("myfass.club", "auth.system", "openfaas-auth-ingress", "nginx", false)
	runYamlTest(parts, "./tmp/openfaas-cloud/templates/ingress/ingress-auth.yml", want, t)
}
func Test_IngressAuthWithTLS(t *testing.T) {
	parts := []string{
		"--set", "global.rootDomain=myfass.club",
		"--set", "tls.enabled=true",
	}

	want := buildIngressAuth("myfass.club", "auth.system", "openfaas-auth-ingress", "nginx", true)
	runYamlTest(parts, "./tmp/openfaas-cloud/templates/ingress/ingress-auth.yml", want, t)
}

func Test_IngressWildcardNoTLS(t *testing.T) {
	parts := []string{
		"--set", "global.rootDomain=myfass.club",
	}
	want := buildIngressAuth("myfass.club", "*", "openfaas-ingress", "nginx", false)
	runYamlTest(parts, "./tmp/openfaas-cloud/templates/ingress/ingress-wildcard.yml", want, t)
}
func Test_IngressWildcardWithTLS(t *testing.T) {
	parts := []string{
		"--set", "global.rootDomain=myfass.club",
		"--set", "tls.enabled=true",
	}

	want := buildIngressAuth("myfass.club", "*", "openfaas-ingress", "nginx", true)
	runYamlTest(parts, "./tmp/openfaas-cloud/templates/ingress/ingress-wildcard.yml", want, t)
}

func Test_IngressWildcardIngressClass(t *testing.T) {

	parts := []string{
		"--set", "global.rootDomain=myfass.club",
		"--set", "global.ingressClass=traefik",
	}

	want := buildIngressAuth("myfass.club", "*", "openfaas-ingress", "traefik", false)
	runYamlTest(parts, "./tmp/openfaas-cloud/templates/ingress/ingress-wildcard.yml", want, t)
}

func Test_IngressAuthIngressClass(t *testing.T) {

	parts := []string{
		"--set", "global.rootDomain=myfass.club",
		"--set", "global.ingressClass=traefik",
	}

	want := buildIngressAuth("myfass.club", "auth.system", "openfaas-auth-ingress", "traefik", false)
	runYamlTest(parts, "./tmp/openfaas-cloud/templates/ingress/ingress-auth.yml", want, t)
}

func buildIngressAuth(hostDomain, prefix, name, ingressClass string, tls bool) YamlSpec {
	annotations := make(map[string]string)
	labels := make(map[string]string)
	backend := make(map[string]string)

	backend["serviceName"] = "edge-router"
	backend["servicePort"] = "8080"

	labels["app"] = "faas-netesd"

	annotations["kubernetes.io/ingress.class"] = ingressClass
	annotations["nginx.ingress.kubernetes.io/limit-connections"] = "20"
	annotations["nginx.ingress.kubernetes.io/limit-rpm"] = "600"

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
			Namespace: 	"openfaas",
			Annotations: annotations,
			Labels:      labels,
		},
		Spec: spec,
	}
}
