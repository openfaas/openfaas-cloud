package function

import (
	"encoding/json"
	"testing"

	"github.com/openfaas/faas-provider/logs"
	"github.com/openfaas/openfaas-cloud/sdk"
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
		logChan := createLogMessageChan(messages)

		want := "some message"

		got := formatLogs(logChan)

		if got != want {
			t.Errorf("want: %q, got: %q", want, got)
		}
	})
}

func Test_formatLogs_multiLogLine(t *testing.T) {
	t.Run("", func(t *testing.T) {
		messages := []string{"some message", "some second message"}
		logChan := createLogMessageChan(messages)

		want := "some message\nsome second message"

		got := formatLogs(logChan)

		if got != want {
			t.Errorf("want: %q, got: %q", want, got)
		}
	})
}

func Test_formatLogs_emptyLogLine(t *testing.T) {
	t.Run("", func(t *testing.T) {
		messages := []string{}
		logString := createLogMessageChan(messages)

		want := ""

		got := formatLogs(logString)

		if got != want {
			t.Errorf("want: %q, got: %q", want, got)
		}
	})
}

func Test_formatLogs_emptyByteInput(t *testing.T) {
	t.Run("", func(t *testing.T) {
		messages := []string{""}
		logChan := createLogMessageChan(messages)

		want := ""

		got := formatLogs(logChan)

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

func createLogMessageChan(messages []string) chan logs.Message {
	logChan := make(chan logs.Message)

	go func() {
		for _, v := range messages {
			logChan <- logs.Message{Text: v}
		}
		close(logChan)
	}()
	return logChan
}
