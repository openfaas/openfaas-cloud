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
	installation_id := "123456"
	os.Setenv("Http_Installation_id", installation_id)
	eventInfo, err := BuildEventFromEnv()
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
	_, err := BuildEventFromEnv()

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
