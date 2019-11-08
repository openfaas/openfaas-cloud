package function

import (
	"encoding/json"
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

// Handle grabs the logs for the fns that are named in the input
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
		return "User is required as POST or querystring i.e. ?user=alexellis."
	}

	c := http.Client{
		Timeout: time.Second * 3,
	}

	gatewayURL := os.Getenv("gateway_url")

	httpReq, _ := http.NewRequest(http.MethodGet, gatewayURL+"system/logs", nil)

	query := url.Values{}
	query.Add("name", function)
	query.Add("follow", "false")
	query.Add("verify_label", fmt.Sprintf("com.openfaas.cloud.git-owner=%s", user))

	addAuthErr := sdk.AddBasicAuth(httpReq)
	if addAuthErr != nil {
		log.Printf("Basic auth error %s", addAuthErr)
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

	if response.StatusCode != http.StatusOK {
		log.Fatalf("unable to query logs, status: %d, message: %s", response.StatusCode, string(bodyBytes))
	}

	formattedLogs := formatLogs(bodyBytes)

	return formattedLogs
}

func formatLogs(msgBody []byte) string {
	if len(msgBody) == 0 {
		return ""
	}
	var b strings.Builder
	for _, line := range strings.Split(strings.TrimSuffix(string(msgBody), "\n"), "\n") {
		data := Message{}
		if err := json.Unmarshal([]byte(line), &data); err != nil {
			log.Fatal(err)
		}
		b.WriteString(data.Text)
	}

	return strings.TrimRight(b.String(), "\n")
}

type Message struct {
	Name      string    `json:"name"`
	Instance  string    `json:"instance"`
	Timestamp time.Time `json:"timestamp"`
	Text      string    `json:"text"`
}
