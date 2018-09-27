package sdk

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/openfaas/faas-provider/auth"
)

const (
	defaultPrivateKeyName  = "private-key"
	defaultSecretMountPath = "/var/openfaas/secrets"
)

// AddBasicAuth to a request by reading secrets when available
func AddBasicAuth(req *http.Request) error {
	if len(os.Getenv("basic_auth")) > 0 && os.Getenv("basic_auth") == "true" {

		reader := auth.ReadBasicAuthFromDisk{}

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

func GetPrivateKeyPath() string {
	// Private key name can be different from the default 'private-key'
	// When providing a different name in the stack.yaml, user need to specify the name
	// in github.yml as `private_key_filename: <user_private_key>`
	privateKeyName := os.Getenv("private_key_filename")

	if privateKeyName == "" {
		privateKeyName = defaultPrivateKeyName
	}

	secretMountPath := os.Getenv("secret_mount_path")

	if secretMountPath == "" {
		secretMountPath = defaultSecretMountPath
	}

	privateKeyPath := filepath.Join(secretMountPath, privateKeyName)

	return privateKeyPath
}
