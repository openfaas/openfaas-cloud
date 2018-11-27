package function

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"

	"github.com/openfaas/faas/gateway/metrics"
)

type Metrics struct {
	Success int `json:"success"`
	Failure int `json:"failure"`
}

// Handle exposes the OpenFaaS instance metrics
func Handle(req []byte) string {
	fnName, err := parseFunctionName()

	if err != nil {
		log.Fatalf("couldn't parse function name from query: %t", err)
	}

	host := os.Getenv("prometheus_host")
	envPort := os.Getenv("prometheus_port")
	port, err := strconv.Atoi(envPort)
	if err != nil {
		log.Fatalf("Could not convert env-var prometheus_port to int. Env-var value: %s. Error: %t", envPort, err)
	}

	metricsQuery := metrics.NewPrometheusQuery(host, port, &http.Client{})
	metricsWindow := parseMetricsWindow()

	fnMetrics, err := getMetrics(fnName, metricsQuery, metricsWindow)
	if err != nil {
		log.Fatalf("Couldn't get metrics from Prometheus for function %s, %t", fnName, err)
	}

	res, err := json.Marshal(fnMetrics)
	if err != nil {
		log.Fatalf("Couldn't marshal json %t", err)
	}
	return string(res)
}

func parseMetricsWindow() string {
	if query, exists := os.LookupEnv("Http_Query"); exists {
		vals, _ := url.ParseQuery(query)

		metricsWindow := vals.Get("metrics_window")

		if len(metricsWindow) > 0 {
			return metricsWindow
		}
	}

	metricsWindow := os.Getenv("metrics_window")

	if metricsWindow == "" {
		metricsWindow = "60m"
	}

	return metricsWindow
}

func parseFunctionName() (functionName string, error error) {
	if query, exists := os.LookupEnv("Http_Query"); exists {
		vals, _ := url.ParseQuery(query)

		functionNameQuery := vals.Get("function")

		if len(functionNameQuery) > 0 {
			return functionNameQuery, nil
		}

		return "", fmt.Errorf("there is no `function` inside env var Http_Query")
	}

	return "", fmt.Errorf("unable to parse Http_Query")
}

func getMetrics(fnName string, metricsQuery metrics.PrometheusQueryFetcher, metricsWindow string) (*Metrics, error) {
	queryValue := fmt.Sprintf(
		`sum(increase(gateway_function_invocation_total{function_name="%s"}[%s])) by (code)`,
		fnName,
		metricsWindow,
	)
	expr := url.QueryEscape(queryValue)

	response, err := metricsQuery.Fetch(expr)
	if err != nil {
		return nil, fmt.Errorf("Failed to get query metrics for function %s, error: %t", fnName, err)
	}

	success := 0
	failure := 0
	for _, v := range response.Data.Result {
		code := v.Metric.Code

		invocations := 0
		invocationsResIndex := len(v.Value) - 1
		if s, ok := v.Value[invocationsResIndex].(string); ok {
			f, err := strconv.ParseFloat(s, 64)
			if err != nil {
				return nil, fmt.Errorf("Couldn't parse invocations count value to float. Value: %s, Error: %t", s, err)
			}
			invocations = int(f)
		}

		if isSuccess(code) {
			success += invocations
		} else {
			failure += invocations
		}
	}

	result := &Metrics{
		Success: success,
		Failure: failure,
	}

	return result, nil
}

func isSuccess(code string) bool {
	statusCode, err := strconv.Atoi(code)

	if err != nil {
		log.Fatalf("couldn't convert code: %s to integeer - received error: %t", code, err)
	}

	if statusCode >= 200 && statusCode < 400 {
		return true
	}

	return false
}
