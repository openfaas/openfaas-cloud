package function

import (
	"encoding/json"
	"os"
	"testing"
)

func TestGetEvent_ReadLabels(t *testing.T) {

	want := map[string]string{
		"com.openfaas.scale": "true",
	}

	val, _ := json.Marshal(want)
	os.Setenv("Http_Labels", string(val))

	eventInfo, err := getEventFromEnv()
	if err != nil {
		t.Errorf(err.Error())
		t.Fail()
	}

	for k, v := range want {
		if _, ok := eventInfo.Labels[k]; !ok {
			t.Errorf("want %s to be present in event.Labels", k)
			continue
		}
		if vv, _ := eventInfo.Labels[k]; vv != v {
			t.Errorf("value of %s, want: %s, got %s", k, v, vv)
		}

	}
}

func TestGetEvent_ReadAnnotations(t *testing.T) {

	want := map[string]string{
		"topic": "function.deployed",
	}

	val, _ := json.Marshal(want)
	os.Setenv("Http_Annotations", string(val))

	eventInfo, err := getEventFromEnv()
	if err != nil {
		t.Errorf(err.Error())
		t.Fail()
	}

	for k, v := range want {
		if _, ok := eventInfo.Annotations[k]; !ok {
			t.Errorf("want %s to be present in event.Labels", k)
			continue
		}
		if vv, _ := eventInfo.Annotations[k]; vv != v {
			t.Errorf("value of %s, want: %s, got %s", k, v, vv)
		}

	}
}

func TestGetEvent_ReadSecrets(t *testing.T) {

	valSt := []string{"s1", "s2"}
	val, _ := json.Marshal(valSt)
	os.Setenv("Http_Secrets", string(val))

	owner := "alexellis"
	os.Setenv("Http_Owner", owner)

	installationID := "123456"
	os.Setenv("Http_Installation_id", installationID)

	eventInfo, err := getEventFromEnv()
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
	_, err := getEventFromEnv()

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
func Test_getMemoryLimit_Swarm(t *testing.T) {
	tests := []struct {
		title         string
		memoryLimit   string
		expectedLimit string
	}{
		{
			title:         "Kubernetes environment variables missing and limit is set",
			memoryLimit:   "30",
			expectedLimit: "30m",
		},
		{
			title:         "Kubernetes environment variables missing and limit is unset",
			memoryLimit:   "",
			expectedLimit: "128m",
		},
	}
	envVar := "function_memory_limit_mb"
	for _, test := range tests {
		t.Run(test.title, func(t *testing.T) {
			os.Setenv(envVar, test.memoryLimit)
			limit := getMemoryLimit()
			if limit != test.expectedLimit {
				t.Errorf("Test failed! Expected: `%v` got: `%v`.", test.expectedLimit, limit)
			}
		})
	}
}

func Test_getMemoryLimit_Kubernetes(t *testing.T) {
	tests := []struct {
		title           string
		exampleVariable string
		memoryLimit     string
		expectedLimit   string
	}{
		{
			title:           "Kubernetes environment variables present and limit is set",
			exampleVariable: "KUBERNETES_SERVICE_PORT",
			memoryLimit:     "30",
			expectedLimit:   "30Mi",
		},
		{
			title:           "Kubernetes environment variables present and limit is unset",
			exampleVariable: "KUBERNETES_SERVICE_PORT",
			memoryLimit:     "",
			expectedLimit:   "128Mi",
		},
	}
	exampleValue := "example_value"
	envVar := "function_memory_limit_mb"

	for _, test := range tests {
		t.Run(test.title, func(t *testing.T) {
			os.Setenv(test.exampleVariable, exampleValue)
			os.Setenv(envVar, test.memoryLimit)
			limit := getMemoryLimit()
			if limit != test.expectedLimit {
				t.Errorf("Test failed! Expected: `%v` got: `%v`.", test.expectedLimit, limit)
			}
		})
	}
}

func Test_getCPULimit_Kubernetes(t *testing.T) {
	tests := []struct {
		title         string
		limitValue    string
		expectedLimit string
		wantAvailable bool
	}{
		{
			title:         "Override test - Kubernetes environment variables present and limit is set",
			limitValue:    "250",
			expectedLimit: "250m",
			wantAvailable: true,
		},
		{
			title:         "Defaults test - Kubernetes environment variables present and limit is unset",
			limitValue:    "",
			expectedLimit: "",
			wantAvailable: false,
		},
	}

	for _, test := range tests {
		t.Run(test.title, func(t *testing.T) {
			os.Setenv("KUBERNETES_SERVICE_PORT", "6443")
			os.Setenv("function_cpu_limit_milli", test.limitValue)

			limit := getCPULimit()
			if limit.Available != test.wantAvailable {
				t.Errorf("Limits not available, want: %v, got: %v", test.wantAvailable, limit.Available)
			}

			if limit.Limit != test.expectedLimit {
				t.Errorf("Limits not correct, want: `%v` got: `%v`.", test.expectedLimit, limit.Limit)
			}
		})
	}
}

func Test_existingVariable_Existent(t *testing.T) {
	tests := []struct {
		title string
		value string
	}{
		{
			title: "Variable exist and set",
			value: "example",
		},
		{
			title: "Variable exist but unset",
			value: "",
		},
	}

	key := "env_var"
	expectedBool := true

	for _, test := range tests {
		t.Run(test.title, func(t *testing.T) {
			os.Setenv(key, test.value)
			_, exists := os.LookupEnv(key)
			//exists := existingVariable(key)
			if exists != expectedBool {
				t.Errorf("Variable existance should be : `%v` got: `%v`", expectedBool, exists)
			}
		})
	}
}

func Test_existingVariable_nonExistent(t *testing.T) {
	t.Run("Variable does not exist", func(t *testing.T) {
		expectedBool := false
		key := "place_holder"
		_, exists := os.LookupEnv(key)

		if exists != expectedBool {
			t.Errorf("Should be:`%v` got:`%v`", expectedBool, exists)
		}
	})
}

func Test_getConfig(t *testing.T) {

	var configOpts = []struct {
		name         string
		value        string
		defaultValue string
		isConfugured bool
	}{
		{
			name:         "scaling_max_limit",
			value:        "",
			defaultValue: "4",
			isConfugured: true,
		},
		{
			name:         "scaling_max_limit",
			value:        "10",
			defaultValue: "4",
			isConfugured: true,
		},
		{
			name:         "random_config",
			value:        "",
			defaultValue: "18",
			isConfugured: false,
		},
	}
	for _, config := range configOpts {
		t.Run(config.name, func(t *testing.T) {
			if config.isConfugured {
				os.Setenv(config.name, config.value)
			}
			value := getConfig(config.name, config.defaultValue)
			want := config.defaultValue
			if len(config.value) > 0 {
				want = config.value
			}

			if value != want {
				t.Errorf("want %s, but got %s", want, value)
			}
		})
	}
}

func Test_buildAnnotations_RemovesNonWhitelisted(t *testing.T) {
	whitelist := []string{"topic"}

	userValues := map[string]string{
		"com.url": "value",
	}

	out := buildAnnotations(whitelist, userValues)

	if _, ok := out["com.url"]; ok {
		t.Fail()
	}

}

func Test_buildAnnotations_AllowsWhitelisted(t *testing.T) {
	whitelist := []string{
		"topic",
		"schedule",
	}

	userValues := map[string]string{
		"topic":    "function.deployed",
		"schedule": "has schedule",
	}

	out := buildAnnotations(whitelist, userValues)

	topicVal, ok := out["topic"]
	if !ok {
		t.Errorf("want user annotation: topic")
		t.Fail()
	}
	if topicVal != userValues["topic"] {
		t.Errorf("want user annotation: topic - got %s, want %s", topicVal, userValues["topic"])
		t.Fail()
	}

	scheduleVal, ok := out["schedule"]
	if !ok {
		t.Errorf("want user annotation: schedule")
		t.Fail()
	}
	if scheduleVal != userValues["schedule"] {
		t.Errorf("want user annotation: schedule - got %s, want %s", scheduleVal, userValues["schedule"])
		t.Fail()
	}
}
