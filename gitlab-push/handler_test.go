package function

import (
	"testing"
)

func Test_checkPublicRepo(t *testing.T) {
	tests := []struct {
		title           string
		visibilityLevel int
		expectedBool    bool
	}{
		{
			title:           "This is visibility level Private, users with access can see it",
			visibilityLevel: PrivateRepo,
			expectedBool:    true,
		},
		{
			title:           "This is visibility level Internal, only logged in users have access",
			visibilityLevel: InternalRepo,
			expectedBool:    true,
		},
		{
			title:           "This is visibility level Public and everyone can see it",
			visibilityLevel: PublicRepo,
			expectedBool:    false,
		},
	}
	for _, test := range tests {
		t.Run(test.title, func(t *testing.T) {
			privateRepo := checkPublicRepo(test.visibilityLevel)
			if privateRepo != test.expectedBool {
				t.Errorf("Unexpected result wanted: `%v` got: `%v`", test.expectedBool, privateRepo)
			}
		})
	}

}
