package function

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/openfaas/openfaas-cloud/sdk"
)

// Handle takes the functions which are built
// by buildshiprun and exposes the function object
// to be consumed by the dashboard so the function
// can be displayed
func Handle(req []byte) string {

	user := string(req)
	if len(user) == 0 {
		if query, exists := os.LookupEnv("Http_Query"); exists {
			vals, _ := url.ParseQuery(query)
			userQuery := vals.Get("user")
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

	httpReq, _ := http.NewRequest(http.MethodGet, gatewayURL+"system/functions", nil)
	addAuthErr := sdk.AddBasicAuth(httpReq)
	if addAuthErr != nil {
		log.Printf("Basic auth error %s", addAuthErr)
	}

	response, err := c.Do(httpReq)
	filtered := []function{}

	if err != nil {
		log.Fatal(err)
	}

	defer response.Body.Close()
	bodyBytes, bErr := ioutil.ReadAll(response.Body)
	if bErr != nil {
		log.Fatal(bErr)
	}

	if response.StatusCode != http.StatusOK {
		log.Fatalf("unable to query functions, status: %d, message: %s", response.StatusCode, string(bodyBytes))
	}

	functions := []function{}
	mErr := json.Unmarshal(bodyBytes, &functions)
	if mErr != nil {
		log.Fatal(mErr)
	}

	for _, fn := range functions {
		for k, v := range fn.Labels {
			if k == (sdk.FunctionLabelPrefix+"git-owner") && strings.EqualFold(v, user) {
				// Hide internal-repo details
				fn.Image = fn.Image[strings.Index(fn.Image, "/")+1:]
				filtered = append(filtered, fn)
			}
		}
	}

	bytesOut, _ := json.Marshal(filtered)
	return string(bytesOut)
}

type function struct {
	Name            string            `json:"name"`
	Image           string            `json:"image"`
	InvocationCount float64           `json:"invocationCount"`
	Replicas        uint64            `json:"replicas"`
	Labels          map[string]string `json:"labels"`
	Annotations     map[string]string `json:"annotations"`
}
