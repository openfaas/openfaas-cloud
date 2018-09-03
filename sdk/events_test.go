package sdk

import (
	"testing"
)

func Test_ServiceName(t *testing.T) {
	values := []struct {
		eventOwner   string
		functionName string
		expectedName string
	}{
		{
			eventOwner:   "ExampleName",
			functionName: "ExampleFunction",
			expectedName: "examplename-ExampleFunction",
		},
		{
			eventOwner:   "examplename",
			functionName: "ExampleFunction",
			expectedName: "examplename-ExampleFunction",
		},
		{
			eventOwner:   "examplename",
			functionName: "examplefunction",
			expectedName: "examplename-examplefunction",
		},
		{
			eventOwner:   "ExampleName",
			functionName: "examplefunction",
			expectedName: "examplename-examplefunction",
		},
	}
	for _, test := range values {
		serviceName := ServiceName(test.eventOwner, test.functionName)
		if serviceName != test.expectedName {
			t.Errorf("Expected name: `%v` got: `%v`", test.expectedName, serviceName)
		}
	}
}
