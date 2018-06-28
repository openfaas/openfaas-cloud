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
)

// Handle a serverless request
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

	request, _ := http.NewRequest(http.MethodGet, gatewayURL+"system/functions", nil)

	response, err := c.Do(request)
	filtered := []function{}

	if err == nil {
		defer response.Body.Close()
		bodyBytes, bErr := ioutil.ReadAll(response.Body)
		if bErr != nil {
			log.Fatal(bErr)
		}

		functions := []function{}
		mErr := json.Unmarshal(bodyBytes, &functions)
		if mErr != nil {
			log.Fatal(mErr)
		}

		for _, fn := range functions {
			for k, v := range fn.Labels {
				if k == "Git-Owner" && v == user {
					// Hide internal-repo details
					fn.Image = fn.Image[strings.Index(fn.Image, "/")+1:]

					filtered = append(filtered, fn)
				}
			}
		}
	}

	bytesOut, _ := json.Marshal(filtered)
	return string(bytesOut)
}

type function struct {
	Name   string            `json:"name"`
	Image  string            `json:"image"`
	Labels map[string]string `json:"labels"`
}
