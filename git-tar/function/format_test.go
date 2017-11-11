package function

import (
	"testing"

	"github.com/openfaas/faas-cli/stack"
)

func TestImageNameSuffix_WithoutLatestTag(t *testing.T) {
	function := &stack.Function{
		Image: "alexellis2/func",
	}

	name := formatImageShaTag("registry:5000", function, "thesha")
	want := "registry:5000/alexellis2/func:latest-thesha"
	if name != want {
		t.Errorf("Want %s, got %s.", want, name)
	}
}

func TestImageNameSuffix_WithSpecificTag(t *testing.T) {
	function := &stack.Function{
		Image: "alexellis2/func:0.1",
	}

	name := formatImageShaTag("registry:5000", function, "thesha")
	want := "registry:5000/alexellis2/func:0.1-thesha"
	if name != want {
		t.Errorf("Want %s, got %s.", want, name)
	}
}
