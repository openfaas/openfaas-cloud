package test

import (
	"testing"
)

func Test_YamlSpecFNNamespace_NoOverrides(t *testing.T) {
	parts := []string{}
	want := buildFnNetworkPolicy("openfaas-fn")
	runYamlTest(parts,"./tmp/openfaas-cloud/templates/network-policy/ns-openfaas-fn-net-policy.yml", want, t)
}

func Test_YamlSpecFNNamespace_Overrides(t *testing.T) {
	parts := []string{
		"--set", "global.functionsNamespace=some-fn-namespace",
	}
	want := buildFnNetworkPolicy("some-fn-namespace")
	runYamlTest(parts,"./tmp/openfaas-cloud/templates/network-policy/ns-openfaas-fn-net-policy.yml", want, t)

}

func Test_CoreNetworkNamespace_NoOverrides(t *testing.T) {
	parts := []string{}
	want := buildCoreNetworkPolicy("openfaas", "openfaas-fn")
	runYamlTest(parts,"./tmp/openfaas-cloud/templates/network-policy/ns-openfaas-net-policy.yml", want, t)
}

func Test_CoreNetworkPolicy_Overrides(t *testing.T) {
	parts := []string{
		"--set", "global.functionsNamespace=some-fn-namespace",
		"--set", "global.coreNamespace=some-namespace",
	}
	want := buildCoreNetworkPolicy("some-namespace", "some-fn-namespace")
	runYamlTest(parts,"./tmp/openfaas-cloud/templates/network-policy/ns-openfaas-net-policy.yml", want, t)
}

func buildCoreNetworkPolicy(coreNamespace, functionNamespace string) YamlSpec {
	podSelector := make(map[string]string)
	emptySelector := make(map[string]string)
	matchLabelsSystem := make(map[string]string)
	matchLabelsFunction := make(map[string]string)

	podSelector["app"] = "nginx-ingress"
	matchLabelsSystem["role"] = "openfaas-system"
	matchLabelsFunction["role"] = functionNamespace

	return YamlSpec{
		ApiVersion: "networking.k8s.io/v1",
		Kind:       "NetworkPolicy",
		Metadata: MetadataItems{
			Name: coreNamespace,
		},
		Spec: Spec{
			PolicyTypes: []string{"Ingress"},
			PodSelector: emptySelector,
			Ingress: []NetworkIngress{{
				From: []NetworkSelectors{
					{
						Namespace: NamespaceSelector{
							MatchLabels: matchLabelsSystem,
						},
					},
					{
						Namespace: NamespaceSelector{
							MatchLabels: matchLabelsFunction,
						},
						Pod: MatchLabelSelector{
							MatchLabels:matchLabelsSystem,
						},
					},
					{
						Namespace: NamespaceSelector{},
						Pod: MatchLabelSelector{
								MatchLabels: podSelector,
							},
						},

					},
				},
			},
		},
	}
}

func buildFnNetworkPolicy(functionNamespace string) YamlSpec {
	podSelector := make(map[string]string)
	matchLabels := make(map[string]string)

	matchLabels["role"] = "openfaas-system"

	return YamlSpec{
		ApiVersion: "networking.k8s.io/v1",
		Kind:       "NetworkPolicy",
		Metadata: MetadataItems{
			Name:      functionNamespace,
			Namespace: functionNamespace,
		},
		Spec: Spec{
			PolicyTypes: []string{"Ingress"},
			PodSelector: podSelector,
			Ingress: []NetworkIngress{{
				From: []NetworkSelectors{
					{
						Namespace: NamespaceSelector{
							MatchLabels: matchLabels,
						},
					},
					{
						Pod: MatchLabelSelector{
							MatchLabels: matchLabels,
						},
					}},
			}},
		},
	}

}
