package test

import "testing"

func Test_SealedSecretsRole_DefaultFlags(t *testing.T) {
	parts := []string{}

	want := makeSecretsRoleBinding("openfaas-fn")
	runYamlTest(parts, "./tmp/openfaas-cloud/templates/sealed-secrets/rbac-import-secrets-role-binding.yml", want, t)
}

func Test_SealedSecretsRole_OveriddenFlags(t *testing.T) {
	parts := []string{
		"--set", "global.functionsNamespace=this-new-ns",
	}

	want := makeSecretsRoleBinding("this-new-ns")
	runYamlTest(parts, "./tmp/openfaas-cloud/templates/sealed-secrets/rbac-import-secrets-role-binding.yml", want, t)
}

func makeSecretsRoleBinding(fnNamespace string) YamlSpec {
	return YamlSpec{
		Kind:       "RoleBinding",
		ApiVersion: "rbac.authorization.k8s.io/v1",
		Metadata: MetadataItems{
			Name:      "manage-sealed-secrets",
			Namespace: fnNamespace,
		},
		Subjects: []Subjects{{
			Kind:      "ServiceAccount",
			Name:      "sealedsecrets-importer-rw",
			Namespace: fnNamespace,
		}},
		RoleRef: map[string]string{
			"kind":     "ClusterRole",
			"name":     "sealedsecrets-importer",
			"apiGroup": "rbac.authorization.k8s.io",
		},
	}
}
