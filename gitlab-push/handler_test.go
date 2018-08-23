package function

import (
	"testing"
)

func Test_tokenMatch(t *testing.T) {
	examples := []struct {
		gitlabToken  string
		loadedToken  string
		expectedBool bool
	}{
		{gitlabToken: "example", loadedToken: "non_example", expectedBool: false},
		{gitlabToken: "example", loadedToken: "example", expectedBool: true},
	}
	for _, test := range examples {
		matching := tokenMatch(test.gitlabToken, test.loadedToken)
		if matching != test.expectedBool {
			t.Errorf("Expected: `%v` got: `%v`", test.expectedBool, matching)
		}
	}
}

func Test_formatPrivateRepo(t *testing.T) {
	tests := []struct {
		title           string
		visibilityLevel int
		expectedBool    bool
	}{
		{
			title:           "This is visibility level Private, users with access can see it",
			visibilityLevel: 00,
			expectedBool:    true,
		},
		{
			title:           "This is visibility level Internal, only logged in users have access",
			visibilityLevel: 10,
			expectedBool:    true,
		},
		{
			title:           "This is visibility level Public and everyone can see it",
			visibilityLevel: 20,
			expectedBool:    false,
		},
	}
	for _, test := range tests {
		t.Run(test.title, func(t *testing.T) {
			privateRepo := formatPrivateRepo(test.visibilityLevel)
			if privateRepo != test.expectedBool {
				t.Errorf("Unexpected result wanted: `%v` got: `%v`", test.expectedBool, privateRepo)
			}
		})
	}

}
