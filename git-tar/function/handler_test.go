package function

import (
	"os"
	"testing"
)

func Test_IsWithinLimitVariableNotSet(t *testing.T) {
	os.Unsetenv("functions_per_repo_limit")

	want := true
	val, _ := isWithinLimit(10)

	if val != want {
		t.Errorf("Limit check failed: want %v, got %v", want, val)
		t.Fail()
	}
}

func Test_IsWithinLimitInvalidValue(t *testing.T) {
	os.Setenv("functions_per_repo_limit", "InvalidValue")

	want := false
	val, _ := isWithinLimit(10)

	if val != want {
		t.Errorf("Limit check failed: want %v, got %v", want, val)
		t.Fail()
	}
}

func Test_IsWithinLimitTooManyFunctions(t *testing.T) {
	os.Setenv("functions_per_repo_limit", "5")

	want := false
	val, _ := isWithinLimit(10)

	if val != want {
		t.Errorf("Limit check failed: want %v, got %v", want, val)
		t.Fail()
	}
}

func Test_IsWithinLimitTooFewFunctions(t *testing.T) {
	os.Setenv("functions_per_repo_limit", "5")

	want := true
	val, _ := isWithinLimit(4)

	if val != want {
		t.Errorf("Limit check failed: want %v, got %v", want, val)
		t.Fail()
	}
}
