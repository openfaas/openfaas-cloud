package sdk

import (
	"fmt"
	"net/http"
	"os"

	"github.com/openfaas/faas/gateway/types"
)

// AddBasicAuth to a request by reading secrets when available
func AddBasicAuth(req *http.Request) error {
	if len(os.Getenv("basic_auth")) > 0 && os.Getenv("basic_auth") == "true" {

		reader := types.ReadBasicAuthFromDisk{}

		if len(os.Getenv("secret_mount_path")) > 0 {
			reader.SecretMountPath = os.Getenv("secret_mount_path")
		}

		credentials, err := reader.Read()

		if err != nil {
			return fmt.Errorf("error with AddBasicAuth %s", err.Error())
		}

		req.SetBasicAuth(credentials.User, credentials.Password)
	}
	return nil
}
