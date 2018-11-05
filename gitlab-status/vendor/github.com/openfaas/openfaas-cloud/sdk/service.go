package sdk

import (
	"fmt"
	"strings"
)

func FormatServiceName(owner, functionName string) string {
	return fmt.Sprintf("%s-%s", strings.ToLower(owner), functionName)
}

func CreateServiceURL(URL, suffix string) string {
	if strings.Contains(URL, suffix) {
		return URL
	}
	columns := strings.Count(URL, ":")
	//columns in URL with port are 2 i.e. http://url:port
	if columns == 2 {
		baseURL := URL[:strings.LastIndex(URL, ":")]
		port := URL[strings.LastIndex(URL, ":"):]
		return fmt.Sprintf("%s.%s%s", baseURL, suffix, port)
	}
	return fmt.Sprintf("%s.%s", URL, suffix)
}

// FormatShortSHA returns a 7-digit SHA
func FormatShortSHA(sha string) string {
	if len(sha) <= 7 {
		return sha
	}
	return sha[:7]
}
