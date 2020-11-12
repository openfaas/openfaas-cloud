package test

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"reflect"
	"testing"
)

func serviceBuilder(svcName, svcType string, srcPort, targetPort int) SvcSpec {
	labelMap := make(map[string]string)

	labelMap["app"] = svcName

	return SvcSpec{
		ApiVersion: "v1",
		Kind:       "Service",
		Metadata: MetadataItems{
			Name:      svcName,
			Namespace: "openfaas",
			Labels:    labelMap,
		},
		Spec: MapSpec{
			Type: svcType,
			Ports: []SvcPorts{{
				Name:       "http",
				Port:       srcPort,
				Protocol:   "TCP",
				TargetPort: targetPort,
			}},
			Selector: labelMap,
		},
	}

}

type SvcSpec struct {
	ApiVersion string        `yaml:"apiVersion"`
	Kind       string        `yaml:"kind"`
	Metadata   MetadataItems `yaml:"metadata"`
	Spec       MapSpec       `yaml:"spec"`
}

type MapSpec struct {
	Type     string            `yaml:"type,omitempty"`
	Ports    []SvcPorts        `yaml:"ports,omitempty"`
	Selector map[string]string `yaml:"selector,omitempty"`
}

type SvcPorts struct {
	Name       string `yaml:"name"`
	Port       int    `yaml:"port"`
	Protocol   string `yaml:"protocol"`
	TargetPort int    `yaml:"targetPort"`
}

func runSvcTest(parts []string, filename string, want SvcSpec, t *testing.T) {
	_, _ = helmRunner(parts...)

	data, err := ioutil.ReadFile(filename)
	if err != nil {
		t.Errorf("got error when reading yaml: %s", err.Error())
	}

	got := SvcSpec{}

	if err = yaml.UnmarshalStrict(data, &got); err != nil {
		t.Errorf("Error reading file %s", err)
		t.Fail()
	}

	if !reflect.DeepEqual(want, got) {
		t.Errorf("got:\n%+v\nbut want:\n%+v", got, want)
		t.Fail()
	}
}

func runYamlTest(parts []string, filename string, want YamlSpec, t *testing.T) {
	_, _ = helmRunner(parts...)

	data, err := ioutil.ReadFile(filename)
	if err != nil {
		t.Errorf("got error when reading yaml: %s", err.Error())
	}

	got := YamlSpec{}

	if err = yaml.UnmarshalStrict(data, &got); err != nil {
		t.Errorf("Error reading file %s", err)
		t.Fail()
	}

	if !reflect.DeepEqual(want, got) {
		t.Errorf("got:\n%+v\nbut want:\n%+v", got, want)
		t.Fail()
	}
}

func runYamlTestNoFileExpected(parts []string, filename string, t *testing.T) {
	os.RemoveAll("./tmp")
	_, _ = helmRunner(parts...)

	_, err := ioutil.ReadFile(filename)

	if err == nil {
		t.Errorf("Expected file not to exist, got a file: %s", filename)
	}

	if !os.IsNotExist(err) {
		t.Errorf("Expected file not to exist, got err: %v", err)
	}
}
