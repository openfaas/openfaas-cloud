package sdk

import (
	"fmt"
	"strings"
)

func FormatServiceName(owner, functionName string) string {
	return fmt.Sprintf("%s-%s", strings.ToLower(owner), functionName)
}
