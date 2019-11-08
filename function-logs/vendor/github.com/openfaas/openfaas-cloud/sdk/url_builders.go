package sdk

import (
	"fmt"
	"net/url"
	"strings"
)

const (
	SystemSubdomain = "system"
)

// FormatEndpointURL takes the gateway_public_url environmental
// variable along with event object to format URL which points to
// the function endpoint
func FormatEndpointURL(gatewayURL string, event *Event) (string, error) {
	systemURL, formatErr := FormatSystemURL(gatewayURL)
	if formatErr != nil {
		return "", fmt.Errorf("error while formattig endpoint URL: %s", formatErr.Error())
	}
	personalURL := strings.Replace(systemURL, SystemSubdomain, event.Owner, -1)

	return fmt.Sprintf("%s/%s", personalURL, event.Service), nil
}

// FormatDashboardURL takes the environmental variable
// gateway_public_url and event object and formats
// the URL to point to the dashboard
func FormatDashboardURL(gatewayURL string, event *Event) (string, error) {
	systemURL, formatErr := FormatSystemURL(gatewayURL)
	if formatErr != nil {
		return "", fmt.Errorf("error while formatting dashboard URL: %s", formatErr.Error())
	}

	return fmt.Sprintf("%s/dashboard/%s", systemURL, event.Owner), nil
}

// GetSubdomain gets the subdomain of the URL
// for example the subdomain of www.o6s.io
// would be www
func GetSubdomain(URL string) (string, error) {
	parsedURL, parseErr := url.Parse(URL)
	if parseErr != nil {
		return "", fmt.Errorf("Unable to parse URL: %s", parseErr.Error())
	}
	subdomain := strings.Split(parsedURL.Host, ".")

	//Host is www.world.org and subdomain would be www aka. 0th element of the slice
	return subdomain[0], nil
}

// FormatSystemURL formats the system URL which points to the
// edge-router with the gateway_public_url environmental variable
func FormatSystemURL(gatewayURL string) (string, error) {
	if strings.HasSuffix(gatewayURL, "/") {
		gatewayURL = strings.TrimSuffix(gatewayURL, "/")
	}
	subdomain, err := GetSubdomain(gatewayURL)
	if err != nil {
		return "", fmt.Errorf("error while geting subdomain for system URL: %s", err)
	}
	systemURL := strings.Replace(gatewayURL, subdomain, SystemSubdomain, -1)
	return systemURL, nil
}

// FormatLogsURL formats the URL where function logs are stored with
// the gateway_public_url environmental variable and event object
func FormatLogsURL(gatewayURL string, event *Event) (string, error) {
	systemURL, formatErr := FormatSystemURL(gatewayURL)
	if formatErr != nil {
		return "", fmt.Errorf("error while formatting logs URL: %s", formatErr.Error())
	}

	return fmt.Sprintf("%s/dashboard/%s/%s/log?repoPath=%s/%s&commitSHA=%s",
		systemURL, event.Owner, event.Service, event.Owner, event.Repository, event.SHA), nil
}
