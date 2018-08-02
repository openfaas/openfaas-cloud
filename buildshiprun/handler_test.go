package function

import (
	"encoding/json"
	"os"
	"testing"
)

func TestBuildURLWithoutPrettyURL_WithSlash(t *testing.T) {
	os.Setenv("gateway_public_url", "http://localhost:8080")
	os.Setenv("gateway_pretty_url", "")

	event := &eventInfo{
		owner:   "alexellis",
		service: "tester",
	}

	val := buildPublicStatusURL("success", event)
	want := "http://localhost:8080/function/alexellis-tester"

	if val != want {
		t.Errorf("building PublicURL: want %s, got %s", want, val)
		t.Fail()
	}
}

func TestBuildURLWithoutPrettyURL_WithOutSlash(t *testing.T) {
	os.Setenv("gateway_public_url", "http://localhost:8080")
	os.Setenv("gateway_pretty_url", "")

	event := &eventInfo{
		owner:   "alexellis",
		service: "tester",
	}

	val := buildPublicStatusURL("success", event)
	want := "http://localhost:8080/function/alexellis-tester"

	if val != want {
		t.Errorf("building PublicURL: want %s, got %s", want, val)
		t.Fail()
	}
}

func TestBuildURLWithPrettyURL(t *testing.T) {
	os.Setenv("gateway_public_url", "http://localhost:8080")
	os.Setenv("gateway_pretty_url", "https://user.openfaas-cloud.com/function")

	event := &eventInfo{
		owner:   "alexellis",
		service: "tester",
	}

	val := buildPublicStatusURL("success", event)
	want := "https://alexellis.openfaas-cloud.com/tester"

	if val != want {
		t.Errorf("building PublicURL: want %s, got %s", want, val)
		t.Fail()
	}
}

func TestBuildURLWithUndefinedStatusGivesOriginalURL(t *testing.T) {

	event := &eventInfo{
		owner:   "alexellis",
		service: "tester",
		url:     "http://original-value.local",
	}

	val := buildPublicStatusURL("not-supported", event)
	want := event.url

	if val != want {
		t.Errorf("building PublicURL: want %s, got %s", want, val)
		t.Fail()
	}
}

func TestGetEvent_ReadSecrets(t *testing.T) {

	valSt := []string{"s1", "s2"}
	val, _ := json.Marshal(valSt)
	os.Setenv("Http_Secrets", string(val))
	owner := "alexellis"
	os.Setenv("Http_Owner", owner)

	eventInfo, err := getEvent()
	if err != nil {
		t.Errorf(err.Error())
		t.Fail()
	}

	expected := []string{owner + "-s1", owner + "-s2"}
	for _, val := range eventInfo.secrets {
		found := false
		for _, expectedVal := range expected {
			if expectedVal == val {
				found = true
			}
		}
		if !found {
			t.Errorf("Wanted secret: %s, didn't find it in list", val)
		}
	}
}

func TestGetEvent_EmptyEnvVars(t *testing.T) {
	_, err := getEvent()

	if err != nil {
		t.Errorf(err.Error())
		t.Fail()
	}
}

func Test_GetImageName(t *testing.T) {

	var imageNameTestcases = []struct {
		Name              string
		PushRepositoryURL string
		RepositoryURL     string
		ImageName         string
		Output            string
	}{
		{
			"Test Docker Hub with user-prefix",
			"docker.io/of-community/",
			"docker.io/of-community/",
			"docker.io/of-community/function-name/",
			"docker.io/of-community/function-name/",
		},
		{
			"Testcase1",
			"registry:5000",
			"127.0.0.1:5000",
			"registry:5000/username/function-name/",
			"127.0.0.1:5000/username/function-name/",
		},
		{
			"Testcase2",
			"registry:31115",
			"127.0.0.1:31115",
			"registry:31115/username/function-name/",
			"127.0.0.1:31115/username/function-name/",
		},
		{
			"Testcase3",
			"registry:31115",
			"127.0.0.1",
			"registry:31115/username/function-name/",
			"127.0.0.1/username/function-name/",
		},
	}

	for _, testcase := range imageNameTestcases {
		t.Run(testcase.Name, func(t *testing.T) {
			output := getImageName(testcase.RepositoryURL, testcase.PushRepositoryURL, testcase.ImageName)
			if output != testcase.Output {
				t.Errorf("%s failed!. got: %s, want: %s", testcase.Name, output, testcase.Output)
			}
		})
	}
}

func Test_ValidImage(t *testing.T) {
	imageNames := map[string]bool{
		"failed to solve: rpc error: code = Unknown desc = exit code 2":   false,
		"failed to solve: rpc error: code = Unknown desc = exit status 2": false,
		"failed to solve:":                                                false,
		"error:":                                                          false,
		"code =":                                                          false,
		"127.0.0.1:5000/someuser/regex_go-regex_py": true,
	}
	for image, expected := range imageNames {
		if validImage(image) != expected {
			t.Errorf("For image %s, got: %v, want: %v", image, !expected, expected)
		}
	}
}

func Test_getReadOnlyRootFS_default(t *testing.T) {
	os.Setenv("readonly_root_filesystem", "1")

	val := getReadOnlyRootFS()
	want := true
	if val != want {
		t.Errorf("want %t, but got %t", want, val)
		t.Fail()
	}
}

func Test_getReadOnlyRootFS_override(t *testing.T) {
	os.Setenv("readonly_root_filesystem", "false")

	val := getReadOnlyRootFS()
	want := false
	if val != want {
		t.Errorf("want %t, but got %t", want, val)
		t.Fail()
	}
}
