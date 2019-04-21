package function

import (
	"errors"
	"os"
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
func Test_filterBranchRef(t *testing.T) {
	tests := []struct {
		title          string
		branchRef      string
		expectedBranch string
	}{
		{
			title:          "Expected format",
			branchRef:      "some/format/master",
			expectedBranch: "master",
		},
		{
			title:          "Unexpected format",
			branchRef:      "dev",
			expectedBranch: "dev",
		},
		{
			title:          "No value",
			branchRef:      "",
			expectedBranch: "",
		},
	}
	for _, test := range tests {
		t.Run(test.title, func(t *testing.T) {
			branch := filterBranchRef(test.branchRef)
			if test.expectedBranch != branch {
				t.Errorf("Expected branch: `%s` got: `%s`", test.expectedBranch, branch)
			}

		})
	}
}
func Test_getBranch(t *testing.T) {
	tests := []struct {
		title          string
		branchInEnv    string
		expectedBranch string
	}{
		{
			title:          "Expected format",
			branchInEnv:    "master",
			expectedBranch: "master",
		},
		{
			title:          "Expected format with spaces",
			branchInEnv:    " master ",
			expectedBranch: "master",
		},
	}
	for _, test := range tests {
		t.Run(test.title, func(t *testing.T) {
			os.Setenv("build_branch", test.branchInEnv)
			branch := getBranch()
			if branch != test.expectedBranch {
				t.Errorf("Value: `%s` not found in expected cases: `%v`", branch, test.expectedBranch)
			}
		})
	}
}

func Test_getBranch_unsetEnv(t *testing.T) {
	expectedBranch := "master"
	branch := getBranch()
	t.Run("Test when env var build_branch is unset", func(t *testing.T) {
		if branch != expectedBranch {
			t.Errorf("Value: `%s` not found in expected cases: `%v`", branch, expectedBranch)
		}
	})
}
func Test_checkBranch(t *testing.T) {
	tests := []struct {
		title         string
		branchesInEnv string
		branchRef     string
		expectedError error
	}{

		{
			title:         "Branch exists in environmental variables",
			branchesInEnv: "master",
			branchRef:     "refs/heads/master",
			expectedError: nil,
		},

		{
			title:         "Branch does not exist in environmental variables",
			branchesInEnv: "staging",
			branchRef:     "/refs/heads/development",
			expectedError: errors.New("refusing to build target branch: development, want branch: staging"),
		},
	}
	for _, test := range tests {
		t.Run(test.title, func(t *testing.T) {
			os.Setenv("build_branch", test.branchesInEnv)
			branchErr := checkBranch(test.branchRef)
			if branchErr != test.expectedError && branchErr != nil {
				if branchErr.Error() != test.expectedError.Error() {
					t.Errorf("Expected error: `%s`, got: `%s`",
						test.expectedError.Error(),
						branchErr.Error())
				}
			}
		})
	}
}

func Test_checkBranch_unsetBranchInEnv(t *testing.T) {
	t.Run("Environmental variable does not exist, only master accepted", func(t *testing.T) {
		os.Unsetenv("build_branch")
		branchRef := "/refs/heads/master"
		branchErr := checkBranch(branchRef)
		if branchErr != nil {
			t.Errorf("Expected error to be nil got: `%s`", branchErr.Error())
		}
	})
}
