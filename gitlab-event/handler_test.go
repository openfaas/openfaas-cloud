package function

import (
	"fmt"
	"testing"
)

func Test_checkSupportedEvents(t *testing.T) {
	tests := []struct {
		title        string
		event        string
		expectedBool bool
	}{
		{
			title:        "Supported `push` event",
			event:        "push",
			expectedBool: true,
		},
		{
			title:        "Supported `project_update` event",
			event:        "project_update",
			expectedBool: true,
		},
		{
			title:        "Supported `project_destroy` event",
			event:        "project_destroy",
			expectedBool: true,
		},
		{
			title:        "Non-supported `repository_update` event",
			event:        "repository_update",
			expectedBool: false,
		},
		{
			title:        "Random string `random words here` event",
			event:        "random words here",
			expectedBool: false,
		},
	}
	for _, test := range tests {
		t.Run(test.title, func(t *testing.T) {
			eventSupported := checkSupportedEvents(test.event)
			if eventSupported != test.expectedBool {
				t.Errorf("expected to be: %v got: %v", test.expectedBool, eventSupported)
			}
		})
	}
}

func Test_getUser(t *testing.T) {
	tests := []struct {
		title             string
		pathWithNamespace string
		expectedName      string
		expectedErr       error
	}{
		{
			title:             "Error is nil and since the string is formated as expected",
			pathWithNamespace: "exampleusername/exampleproject",
			expectedName:      "exampleusername",
			expectedErr:       nil,
		},
		{
			title:             "Error is not nil since the string is not formatted as expected",
			pathWithNamespace: "exampleusername:exampleproject",
			expectedName:      "",
			expectedErr:       fmt.Errorf("un-proper format of the variable possible out of range error"),
		},
		{
			title:             "Case when field is empty",
			pathWithNamespace: "",
			expectedName:      "",
			expectedErr:       fmt.Errorf("un-proper format of the variable possible out of range error"),
		},
	}
	for _, test := range tests {
		t.Run(test.title, func(t *testing.T) {
			username, userErr := getUser(test.pathWithNamespace)
			if username != test.expectedName {
				t.Errorf("expected name: `%s` got: `%s`", test.expectedName, username)
			}
			if userErr == nil {
				if userErr != test.expectedErr {
					t.Errorf("expected error: `%s` got: `%s`", test.expectedErr, userErr)
				}
			} else {
				if userErr.Error() != test.expectedErr.Error() {
					t.Errorf("expected error: `%s` got: `%s`", test.expectedErr, userErr)
				}
			}
		})
	}
}
