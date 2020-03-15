package function

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	faasSDK "github.com/openfaas/faas-cli/proxy"
	"github.com/openfaas/faas-provider/logs"
	"github.com/openfaas/faas-provider/types"
	"github.com/openfaas/openfaas-cloud/sdk"
)

type FaaSAuth struct{}

func (auth *FaaSAuth) Set(req *http.Request) error {
	return sdk.AddBasicAuth(req)
}

var (
	timeout   = 5 * time.Second
	namespace = ""
)

// Handle grabs the logs for the fn that is named in the input
func Handle(req []byte) string {

	user := string(req)
	var function string
	if len(user) == 0 {
		if query, exists := os.LookupEnv("Http_Query"); exists {
			vals, _ := url.ParseQuery(query)
			userQuery := vals.Get("user")
			function = vals.Get("function")
			if len(userQuery) > 0 {
				user = userQuery
			}
		}
	}

	if len(user) == 0 {
		log.Fatalf("User is required as POST or querystring i.e. ?user=alexellis.")
	}

	gatewayURL := os.Getenv("gateway_url")

	allowed, err := isUserFunction(function, gatewayURL, user)

	if err != nil {
		log.Fatalf("there was an error requesting the function %q: %s", function, err.Error())
	}

	if !allowed {
		log.Fatalf("requested function %q could not be found or you are not allowed to access it", function)
	}

	client := faasSDK.NewClient(&FaaSAuth{}, gatewayURL, nil, &timeout)

	ctx := context.Background()

	formattedLogs, fmtErr := getFormattedLogs(*client, ctx, function)

	if fmtErr != nil {
		log.Fatalf("there was an error formatting logs for the function %q, %s", function, fmtErr)
	}
	return formattedLogs
}

func getFormattedLogs(client faasSDK.Client, ctx context.Context, function string) (string, error) {

	if len(function) == 0 {
		return "", errors.New("function name was empty, please provide a valid function name")
	}
	timeSince := time.Now().Add(-1 * time.Minute * 30)
	logRequest := logs.Request{Name: function, Since: &timeSince, Follow: false}

	logChan, err := client.GetLogs(ctx, logRequest)
	if err != nil {
		return "", errors.New(fmt.Sprintf("unable to query logs, message: %s", err.Error()))
	}

	formattedLogs := formatLogs(logChan)

	return formattedLogs, nil
}

func isUserFunction(function string, gatewayURL string, user string) (bool, error) {

	if len(user) == 0 {
		return false, errors.New("user is not set, user must be set for us to find logs")
	}

	bytesIn := []byte("")

	nameQuery := []string{fmt.Sprintf("user=%s", user)}

	resBytes, err := faasSDK.InvokeFunction(gatewayURL, "list-functions", &bytesIn, "", nameQuery, nil, false, "POST", true, "")

	if err != nil {
		return false, errors.New(fmt.Sprintf("unable to query functions list message: %s", err.Error()))
	}

	res, err := functionInResponse(*resBytes, function, user)
	if err != nil {
		return false, err
	}

	return res, nil
}

func formatLogs(logChan <-chan logs.Message) string {
	var b strings.Builder
	for v := range logChan {
		for _, line := range strings.Split(strings.TrimSuffix(v.Text, "\n"), "\n") {
			b.WriteString(line + "\n")
		}
	}
	return strings.TrimRight(b.String(), "\n")
}

func functionInResponse(bodyBytes []byte, function string, owner string) (bool, error) {
	functions := []types.FunctionStatus{}
	mErr := json.Unmarshal(bodyBytes, &functions)
	if mErr != nil {
		return false, mErr
	}

	for _, fn := range functions {
		if fn.Name == function {
			return (*fn.Labels)["com.openfaas.cloud.git-owner"] == owner, nil
		}
	}
	return false, nil
}
