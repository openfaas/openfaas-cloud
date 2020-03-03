package test

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"reflect"
	"testing"
)

func Test_CoreOFBuilderSvc(t *testing.T) {
	parts := []string{}

	want := serviceBuilder("of-builder", "ClusterIP", 8080, 8080)
	_, _ = helmRunner(parts...)

	data, err := ioutil.ReadFile("./tmp/openfaas-cloud/templates/ofc-core/of-builder-svc.yaml")
	if err != nil {
		t.Errorf("got error when reading yaml: %s", err.Error())
	}
	got := SvcSpec{}

	if err = yaml.UnmarshalStrict(data, &got); err != nil {
		t.Errorf("Error reading file %s", err)
		t.Fail()
	}

	if !reflect.DeepEqual(want, got) {
		t.Errorf("got:\n%q\nbut want:\n%q", got, want)
		t.Fail()
	}
}
