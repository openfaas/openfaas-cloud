package function

import (
	"encoding/json"
	"fmt"
	"github.com/openfaas/openfaas-cloud/sdk"
	"strings"
	"testing"
	"time"
)

func Test_functionInResponse_fnIsInResponse(t *testing.T) {
	t.Run("input contains the function that we requested and it has the user's git-owner label", func(t *testing.T) {
		labels := make(map[string]string)
		labels["com.openfaas.cloud.git-owner"] = "my-username"
		inputBytes := []byte(createFnListJson("my-function", labels))
		got, err := functionInResponse(inputBytes, "my-function", "my-username")

		if err != nil {
			t.Errorf("an error was thrown when we didnt expect it, %s", err)
		}

		if got == false {
			t.Error("function was not found in input")

		}
	})
}

func Test_functionInResponse_fnNotInResponse(t *testing.T) {
	t.Run("input does not contain the function that we requested", func(t *testing.T) {
		labels := make(map[string]string)
		labels["com.openfaas.cloud.git-owner"] = "my-username"

		inputBytes := []byte(createFnListJson("my-function-or-is-it", labels))
		got, err := functionInResponse(inputBytes, "my-function", "my-username")

		if err != nil {
			t.Errorf("an error was thrown when we didnt expect it, %s", err)
		}

		if got == true {
			t.Error("function was found in input, it shouldn't have been")

		}
	})
}

func Test_functionInResponse_fnDoesNotHaveCorrectLabelResponse(t *testing.T) {
	t.Run("input contains the function that we requested and it does not have the user's git-owner label", func(t *testing.T) {
		labels := make(map[string]string)
		labels["com.openfaas.cloud.git-owner"] = "not-my-username"
		inputBytes := []byte(createFnListJson("my-function", labels))
		got, err := functionInResponse(inputBytes, "my-function", "my-username")

		if err != nil {
			t.Errorf("an error was thrown when we didnt expect it, %s", err)
		}
		if got == true {
			t.Error("function was found in input, it shouldn't have been as it didnt contain our user in the label")
		}
	})
}
func Test_functionInResponse_fnDoesNotHaveExpectedLabelResponse(t *testing.T) {
	t.Run("input contains the function that we requested and it does not have any git-owner label", func(t *testing.T) {
		labels := make(map[string]string)
		inputBytes := []byte(createFnListJson("my-function", labels))
		got, err := functionInResponse(inputBytes, "my-function", "my-username")

		if err != nil {
			t.Errorf("an error was thrown when we didnt expect it, %s", err)
		}
		if got == true {
			t.Error("function was found in input, it shouldn't have been as it didnt contain our user in the label")
		}
	})
}

func Test_functionInResponse_cantParseResponse(t *testing.T) {
	t.Run("bytes cant be parsed into the function struct (not json)", func(t *testing.T) {
		inputBytes := []byte("some non json stuff")
		_, err := functionInResponse(inputBytes, "my-function", "my-username")

		if err == nil {
			t.Errorf("an error was not thrown when we expect it, %s", err)
		}

	})
}

func Test_functionInResponse_cantParseEmptyResponse(t *testing.T) {
	t.Run("bytes cant be parsed into the function struct (not json)", func(t *testing.T) {
		inputBytes := []byte("")
		_, err := functionInResponse(inputBytes, "my-function", "my-username")

		if err == nil {
			t.Errorf("an error was not thrown when we expect it, %s", err)
		}
	})
}

func Test_formatLogs_singleLogLine(t *testing.T) {
	t.Run("", func(t *testing.T) {
		messages := []string{"some message"}
		logString := createLogMessage(messages)

		want := "some message"

		got, err := formatLogs([]byte(logString))

		if err != nil {
			t.Errorf("an error was thrown when we didnt expect it, %s", err)
		}
		if got != want {
			t.Errorf("want: %q, got: %q", want, got)
		}
	})
}

func Test_formatLogs_multiLogLine(t *testing.T) {
	t.Run("", func(t *testing.T) {
		messages := []string{"some message", "some second message"}
		logString := createLogMessage(messages)

		want := "some messagesome second message"

		got, err := formatLogs([]byte(logString))

		if err != nil {
			t.Errorf("an error was thrown when we didnt expect it, %s", err)
		}
		if got != want {
			t.Errorf("want: %q, got: %q", want, got)
		}
	})
}

func Test_formatLogs_invalidLogLine(t *testing.T) {
	t.Run("", func(t *testing.T) {
		_, err := formatLogs([]byte("spfnsodufbnwiubfiwybca ciabcai"))

		if err == nil {
			t.Error("an error was not thrown when we expect it")
		}
	})
}

func Test_formatLogs_emptyLogLine(t *testing.T) {
	t.Run("", func(t *testing.T) {
		messages := []string{""}
		logString := createLogMessage(messages)

		want := ""

		got, err := formatLogs([]byte(logString))

		if err != nil {
			t.Errorf("an error was thrown when we didnt expect it, %s", err)
		}
		if got != want {
			t.Errorf("want: %q, got: %q", want, got)
		}
	})
}

func Test_formatLogs_emptyByteInput(t *testing.T) {
	t.Run("", func(t *testing.T) {
		want := ""
		got, err := formatLogs([]byte(""))

		if err != nil {
			t.Errorf("an error was thrown when we didnt expect it, %s", err)
		}
		if got != want {
			t.Errorf("want: %q, got: %q", want, got)
		}
	})
}

// test utility fn
func createFnListJson(name string, labels map[string]string) string {
	var functionObj []sdk.Function
	functionObj = append(functionObj, sdk.Function{
		Name:            name,
		Image:           "this-is-not-important",
		InvocationCount: 0,
		Replicas:        0,
		Labels:          labels,
		Annotations:     nil,
	})
	jsonStr, _ := json.Marshal(functionObj)
	return string(jsonStr)

}

func createLogMessage(messages []string) string {

	var b strings.Builder

	for _, msg := range messages {
		var logObj Message
		logObj = Message{
			Name:      "somename",
			Instance:  "somepod",
			Timestamp: time.Time{},
			Text:      msg,
		}

		jsonStr, _ := json.Marshal(logObj)
		b.WriteString(fmt.Sprintf("%s\n", jsonStr))
	}

	return b.String()
}
