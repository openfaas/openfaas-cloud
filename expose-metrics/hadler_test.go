package function

import (
	"encoding/json"
	"testing"

	"github.com/openfaas/faas/gateway/metrics"
)

func Test_parseFunctionName(t *testing.T) {

	query := "http://127.0.0.1:8080/function/fnName"
	val, _ := parseFunctionName(query)
	want := "fnName"
	if val != want {
		t.Errorf("Expected to parse function name %s, but got %s", want, val)
	}

	query = "http://127.0.0.1:8080/async-function/fnName"
	val, _ = parseFunctionName(query)
	if val != want {
		t.Errorf("Expected to parse function name %s, but got %s", want, val)
	}

	query = "http://127.0.0.1:8080/system/function/fnName"
	val, _ = parseFunctionName(query)
	if val != want {
		t.Errorf("Expected to parse function name %s, but got %s", want, val)
	}

	query = "http://127.0.0.1:8080/system/async-function/fnName"
	val, _ = parseFunctionName(query)
	if val != want {
		t.Errorf("Expected to parse function name %s, but got %s", want, val)
	}

	query = "http://127.0.0.1:8080/async-function/fnName/path/after"
	val, _ = parseFunctionName(query)
	if val != want {
		t.Errorf("Expected to parse function name %s, but got %s", want, val)
	}

	query = "http://127.0.0.1:8080/function/fnName/path/after"
	val, _ = parseFunctionName(query)
	if val != want {
		t.Errorf("Expected to parse function name %s, but got %s", want, val)
	}

	query = "http://127.0.0.1:8080/fnName"
	want = ""
	val, _ = parseFunctionName(query)
	if val != want {
		t.Errorf("Expected to parse function name %s, but got %s", want, val)
	}

}

type FakePrometheusQueryFetcher struct {
}

func (q FakePrometheusQueryFetcher) Fetch(query string) (*metrics.VectorQueryResponse, error) {
	val := []byte(`{"Data":{"Result":[{"Metric":{"code":"200","function_name":""},"value":[1536944521.415,"6.0068060449603875"]},{"Metric":{"code":"500","function_name":""},"value":[1536944521.415,"5.005671704133656"]}]}}`)
	queryRes := metrics.VectorQueryResponse{}
	err := json.Unmarshal(val, &queryRes)
	return &queryRes, err
}

func makeFakePrometheusQueryFetcher() FakePrometheusQueryFetcher {
	return FakePrometheusQueryFetcher{}
}

func Test_getMetrics(t *testing.T) {
	fakeQuery := makeFakePrometheusQueryFetcher()
	fnName := "testFunc"

	metrics, _ := getMetrics(fnName, fakeQuery)
	val, _ := json.Marshal(metrics)
	want := `{"success":6,"failure":5}`
	if string(val) != want {
		t.Errorf("Expected %s, but got %s", want, val)
	}
}
