package function

import (
	"encoding/json"
	"os"
	"testing"
)

func TestGetEvent_ReadSecrets(t *testing.T) {

	valSt := []string{"s1", "s2"}
	val, _ := json.Marshal(valSt)
	os.Setenv("Http_Secrets", string(val))

	owner := "alexellis"
	os.Setenv("Http_Owner", owner)

	installationID := "123456"
	os.Setenv("Http_Installation_id", installationID)

	eventInfo, err := getEvent()
	if err != nil {
		t.Errorf(err.Error())
		t.Fail()
	}

	expected := []string{owner + "-s1", owner + "-s2"}
	for _, val := range eventInfo.Secrets {
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

		// error cases
		"failed to solve: rpc error: code = Unknown desc = exit code 2":   false,
		"failed to solve: rpc error: code = Unknown desc = exit status 2": false,
		"failed to solve:":                                                false,
		"error:":                                                          false,
		"code =":                                                          false,
		"-1":                                                              false,
		"":                                                                false,
		" ":                                                               false,

		// "docker-registry:5000/admin/alexellis-sofia-test1-go-world:0.1-374448ba4d75bcf49611525a5b2448d9c3d0bf28": true,
		// url (with/without tag)
		"docker.io/ofcommunity/someuser/repo-name-function_name":                                                 true,
		"docker.io/ofcommunity/someuser/repo-name-function_name:latest":                                          true,
		"docker.io/ofcommunity/someuser/repo-name-function_name:latest-7f7ec13d12b1397408e54b79686d43e41974bfa0": true,

		// url with port (with/without tag)
		"docker.io:80/ofcommunity/someuser/repo-name-function_name":                                                 true,
		"docker.io:80/ofcommunity/someuser/repo-name-function_name:latest":                                          true,
		"docker.io:80/ofcommunity/someuser/repo-name-function_name:latest-7f7ec13d12b1397408e54b79686d43e41974bfa0": true,

		// url with ip (with/without tag)
		"127.0.0.1/someuser/repo-name-function_name":                                                 true,
		"127.0.0.1/someuser/repo-name-function_name:latest":                                          true,
		"127.0.0.1/someuser/repo-name-function_name:latest-7f7ec13d12b1397408e54b79686d43e41974bfa0": true,

		// url with ip and port (with/without tag)
		"127.0.0.1:5000/someuser/repo-name-function_name":                                                 true,
		"127.0.0.1:5000/someuser/repo-name-function_name:latest":                                          true,
		"127.0.0.1:5000/someuser/repo-name-function_name:latest-7f7ec13d12b1397408e54b79686d43e41974bfa0": true,

		// docker user specific (with/without tag)
		"someuser/repo-name-function_name":                                                 true,
		"someuser/repo-name-function_name:latest":                                          true,
		"someuser/repo-name-function_name:latest-7f7ec13d12b1397408e54b79686d43e41974bfa0": true,

		// open faas cloud function name (with/without tag)
		"repo-name-function_name":                                                 true,
		"repo-name-function_name:latest":                                          true,
		"repo-name-function_name:latest-7f7ec13d12b1397408e54b79686d43e41974bfa0": true,

		// simple function name (with/without tag)
		"function_name":                                                 true,
		"function_name:latest":                                          true,
		"function_name:latest-7f7ec13d12b1397408e54b79686d43e41974bfa0": true,
	}

	for image, want := range imageNames {
		got := validImage(image)
		if got != want {
			t.Errorf("Validating image %s - want: %v, got: %v", image, want, got)
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
