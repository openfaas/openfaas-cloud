package test

import "testing"

func Test_RBACImportSecretsServiceAccount_DefaultNS(t *testing.T) {
	parts := []string{}

	want := makeSealedSecretsSA("openfaas-fn")
	runYamlTest(parts, "./tmp/openfaas-cloud/templates/sealed-secrets/rbac-import-secrets-service-account.yml", want, t)
}

func Test_RBACImportSecretsServiceAccount_CustomNS(t *testing.T) {
	parts := []string{
		"--set", "global.functionsNamespace=some-other-ns",
	}

	want := makeSealedSecretsSA("some-other-ns")
	runYamlTest(parts, "./tmp/openfaas-cloud/templates/sealed-secrets/rbac-import-secrets-service-account.yml", want, t)
}

func makeSealedSecretsSA(fnNamespace string) YamlSpec {
	return YamlSpec{
		ApiVersion: "v1",
		Kind:       "ServiceAccount",
		Metadata: MetadataItems{
			Name:      "sealedsecrets-importer-rw",
			Namespace: fnNamespace,
			Labels:    map[string]string{"app": "openfaas"},
		},
	}
}
