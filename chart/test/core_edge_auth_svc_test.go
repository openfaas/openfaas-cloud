package test

import (
	"testing"
)

func Test_CoreEdgeAuthSvc(t *testing.T) {
	parts := []string{}

	want := serviceBuilder("edge-auth", "ClusterIP", 8080, 8080)
	runSvcTest(parts, "./tmp/openfaas-cloud/templates/ofc-core/edge-auth-svc.yaml", want, t)
	_, _ = helmRunner(parts...)
}
