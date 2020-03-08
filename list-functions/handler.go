package function

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/openfaas/faas-provider/types"
	"github.com/openfaas/openfaas-cloud/sdk"

	faasSDK "github.com/openfaas/faas-cli/proxy"
)

type FaaSAuth struct{}

func (auth *FaaSAuth) Set(req *http.Request) error {
	return sdk.AddBasicAuth(req)
}

var (
	timeout   = 3 * time.Second
	namespace = ""
)

// Handle takes the functions which are built
// by buildshiprun and exposes the function object
// to be consumed by the dashboard so the function
// can be displayed
func Handle(req []byte) string {

	gatewayURL := os.Getenv("gateway_url")

	client := faasSDK.NewClient(&FaaSAuth{}, gatewayURL, nil, &timeout)

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

	functions, err := client.ListFunctions(context.Background(), namespace)
	if err != nil {
		log.Fatal(err)
	}

	filtered := []types.FunctionStatus{}
	for _, fn := range functions {
		for k, v := range *fn.Labels {
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
