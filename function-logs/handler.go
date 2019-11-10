package function

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/openfaas/openfaas-cloud/sdk"
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
		log.Fatalf("there was an error requesting the function %q", function)
	}

	if !allowed {
		log.Fatalf("requested function %q could not be found or you are not allowed to access it", function)
	}

	formattedLogs, fmtErr := getFormattedLogs(gatewayURL, function)

	if fmtErr != nil {
		log.Fatalf("there was an error formatting logs for the function %q, %s", function, fmtErr)
	}
	return formattedLogs
}

func getFormattedLogs(gatewayURL string, function string) (string, error) {

	if len(function) == 0 {
		return "", errors.New("function name was empty, please provide a valid function name")
	}
	queryParams := make(map[string]string)

	queryParams["name"] = function
	queryParams["follow"] = "false"

	response, bodyBytes := makeGatewayHttpReq(gatewayURL+"/system/logs", queryParams)

	if response.StatusCode != http.StatusOK {
		return "", errors.New(fmt.Sprintf("unable to query logs, status: %d, message: %s", response.StatusCode, string(bodyBytes)))
	}

	formattedLogs, formatErr := formatLogs(bodyBytes)

	if formatErr != nil {
		return "", formatErr
	}
	return formattedLogs, nil
}

func isUserFunction(function string, gatewayURL string, user string) (bool, error) {
	queryParams := make(map[string]string)
	queryParams["user"] = user

	if len(user) == 0 {
		return false, errors.New("user is not set, user must be set for us to find logs")
	}

	response, bodyBytes := makeGatewayHttpReq(gatewayURL+"/function/list-functions", queryParams)

	if response.StatusCode != http.StatusOK {
		return false, errors.New(fmt.Sprintf("unable to query functions list, status: %d, message: %s", response.StatusCode, string(bodyBytes)))
	}

	res, err := functionInResponse(bodyBytes, function, user)
	if err != nil {
		return false, err
	}

	return res, nil
}

func formatLogs(msgBody []byte) (string, error) {
	if len(msgBody) == 0 {
		return "", nil
	}
	var b strings.Builder
	for _, line := range strings.Split(strings.TrimSuffix(string(msgBody), "\n"), "\n") {
		data := Message{}
		if err := json.Unmarshal([]byte(line), &data); err != nil {
			return "", err
		}
		b.WriteString(data.Text)
	}

	return strings.TrimRight(b.String(), "\n"), nil
}

func functionInResponse(bodyBytes []byte, function string, owner string) (bool, error) {
	functions := []sdk.Function{}
	mErr := json.Unmarshal(bodyBytes, &functions)
	if mErr != nil {
		return false, mErr
	}

	for _, fn := range functions {
		if fn.Name == function {
			return fn.Labels["com.openfaas.cloud.git-owner"] == owner, nil
		}
	}
	return false, nil
}

func makeGatewayHttpReq(URL string, queryParams map[string]string) (*http.Response, []byte) {
	c := http.Client{
		Timeout: time.Second * 3,
	}

	httpReq, _ := http.NewRequest(http.MethodGet, URL, nil)

	query := url.Values{}

	for key, value := range queryParams {
		query.Add(key, value)
	}

	addAuthErr := sdk.AddBasicAuth(httpReq)
	if addAuthErr != nil {
		log.Fatalf("Basic auth error %s", addAuthErr)
	}

	httpReq.URL.RawQuery = query.Encode()

	response, err := c.Do(httpReq)

	if err != nil {
		log.Fatal(err)
	}
	defer response.Body.Close()
	bodyBytes, bErr := ioutil.ReadAll(response.Body)
	if bErr != nil {
		log.Fatal(bErr)
	}

	return response, bodyBytes
}

type Message struct {
	Name      string    `json:"name"`
	Instance  string    `json:"instance"`
	Timestamp time.Time `json:"timestamp"`
	Text      string    `json:"text"`
}
