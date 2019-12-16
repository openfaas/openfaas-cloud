package function

import (
	"github.com/openfaas/openfaas-cloud/sdk"
	"os"
	"strings"
)

func buildPrettyURL(url string, success, isStack bool, event *sdk.Event) string {
	if len(url) == 0 {
		return ""
	}
	if success {
		if isStack {
			urlOut := strings.Replace(url, "user", "system", 1)
			return replaceFunctionSuffix(urlOut, "dashboard") + "/" + event.Owner
		} else {
			urlOut := strings.Replace(url, "user", strings.ToLower(event.Owner), 1)
			return replaceFunctionSuffix(urlOut, event.Service)
		}
	}
	urlOut := strings.Replace(url, "user", "system", 1)

	if isStack {
		return replaceFunctionSuffix(urlOut, "dashboard") + "/" + event.Owner
	}
	return replaceFunctionSuffix(urlOut, "dashboard") + "/" + event.Owner + "/" + event.Service + "/build-log?repoPath=" + event.Owner + "/" + event.Repository + "&commitSHA=" + event.SHA
}

func buildPublicURL(url, owner, service string, success, isStack bool) string {

	if strings.HasSuffix(url, "/") == false {
		url = url + "/"
	}

	if success && !isStack {
		serviceValue := sdk.FormatServiceName(owner, service)
		url = url + "function/" + serviceValue
	} else {
		url = url + "function/system-dashboard"
	}
	return url
}

func buildPublicStatusURL(status, statusContext string, event *sdk.Event) string {
	url := event.URL
	isStack := statusContext == sdk.StackContext
	isSuccess := status == sdk.StatusSuccess
	publicURL := buildPublicURL(os.Getenv("gateway_public_url"), event.Owner, event.Service, isSuccess, isStack)
	gatewayPrettyURL := buildPrettyURL(os.Getenv("gateway_pretty_url"), isSuccess, isStack, event)

	if status == sdk.StatusSuccess {
		if len(gatewayPrettyURL) > 0 {
			return gatewayPrettyURL
		} else if len(publicURL) > 0 {
			return publicURL
		}
	} else if status == sdk.StatusFailure {
		if len(gatewayPrettyURL) > 0 {
			url = gatewayPrettyURL
		} else if len(publicURL) > 0 {
			url = publicURL
		}
	}
	return url

}

func replaceFunctionSuffix(url, newSuffix string) string {
	if strings.HasSuffix(url, "function/") {
		url = strings.TrimSuffix(url, "function/")
	} else {
		url = strings.TrimSuffix(url, "function")
	}

	if strings.HasSuffix(url, "/") {
		return url + newSuffix
	}
	return url + "/" + newSuffix
}
