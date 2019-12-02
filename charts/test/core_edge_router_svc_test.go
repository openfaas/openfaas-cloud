package test

import (
	"testing"
)

func Test_CoreEdgeRouterSvc(t *testing.T) {
	var parts []string
	want := serviceBuilder("edge-router","ClusterIP", 8080, 8080)
	runSvcTest(parts, "./tmp/openfaas-cloud/templates/ofc-core/edge-router-svc.yaml", want, t)
}
