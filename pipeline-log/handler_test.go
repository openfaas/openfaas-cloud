package function

import (
	"os"
	"testing"

	"github.com/openfaas/openfaas-cloud/sdk"
)

func Test_getPath(t *testing.T) {
	got := getPath("pipeline", &sdk.PipelineLog{
		RepoPath:  "alexellis/super-cake",
		CommitSHA: "af6db",
		Function:  "slack-fn",
	})
	want := "pipeline/alexellis/super-cake/af6db/slack-fn/build.log"
	if got != want {
		t.Errorf("got: %s, but want: %s", got, want)
		t.Fail()
	}
}

func Test_tlsEnabled(t *testing.T) {
	connection := []struct {
		title         string
		value         string
		expectedValue bool
	}{
		{
			title:         "Connection enabled",
			value:         "true",
			expectedValue: true,
		},
		{
			title:         "Connection enabled",
			value:         "1",
			expectedValue: true,
		},
		{
			title:         "Connection disabled",
			value:         "every other case is disabled",
			expectedValue: false,
		},
	}
	for _, test := range connection {
		os.Setenv("s3_tls", test.value)
		secured := tlsEnabled()
		if secured != test.expectedValue {
			t.Fatalf("Expected: %v got :%v.\n", test.expectedValue, secured)
		}
	}
}

func Test_bucketName(t *testing.T) {
	bucketNames := []struct {
		title         string
		value         string
		expectedValue string
	}{
		{
			title:         "Bucket name env-var is present",
			value:         "example-name",
			expectedValue: "example-name",
		},
		{
			title:         "Bucket name when env-var does not exist/unset",
			value:         "",
			expectedValue: "pipeline",
		},
	}
	for _, test := range bucketNames {
		os.Setenv("s3_bucket", test.value)
		bucketName := bucketName()
		if bucketName != test.expectedValue {
			t.Errorf("Unexpected bucket name wanted: `%v` got: `%v`", test.expectedValue, bucketName)
		}
	}
}
func Test_regionName(t *testing.T) {
	values := []struct {
		title         string
		value         string
		expectedValue string
	}{
		{
			title:         "Region name is set",
			value:         "eu-west-3",
			expectedValue: "eu-west-3",
		},
		{
			title:         "Region name unset",
			value:         "",
			expectedValue: "us-east-1",
		},
	}
	for _, test := range values {
		os.Setenv("s3_region", test.value)
		regionName := regionName()
		if regionName != test.expectedValue {
			t.Errorf("Expected region name: `%v` got: `%v`\n", test.expectedValue, regionName)
		}
	}

}
