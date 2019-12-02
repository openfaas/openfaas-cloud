package test

import "testing"

func Test_RBACImportSecretsClusterRole(t *testing.T) {
	parts := []string{
	}

	want := makeSealedSecretsClusterRole()
	runYamlTest(parts, "./tmp/openfaas-cloud/templates/sealed-secrets/rbac-import-secrets-cluster-role.yml", want, t)
}


func makeSealedSecretsClusterRole() YamlSpec {
	return YamlSpec{
		ApiVersion: "rbac.authorization.k8s.io/v1",
		Kind:       "ClusterRole",
		Metadata:   MetadataItems{
			Name:        "sealedsecrets-importer",
		},
		Rules: []Rules{{
			ApiGroups: []string{"bitnami.com"},
			Resources: []string{"sealedsecrets"},
			Verbs:     []string{"get", "create", "update"},
		}},
	}
}