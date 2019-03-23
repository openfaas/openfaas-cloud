package function

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
)

const (
	testSecretKeyFile = "github-secret-key"
	privateKeyFile    = "github-private-key"
)

type Config struct {
	SecretKey     string
	PrivateKey    string
	ApplicationID string
}

func NewConfig() (Config, error) {
	config := Config{}

	keyPath, pathErr := getSecretPath()
	if pathErr != nil {
		return config, pathErr
	}

	secretKeyBytes, readErr := ioutil.ReadFile(path.Join(keyPath, testSecretKeyFile))

	if readErr != nil {
		msg := fmt.Errorf("unable to read GitHub symmetrical secret: %s, error: %s",
			keyPath+testSecretKeyFile, readErr)
		return config, msg
	}

	secretKeyBytes = getFirstLine(secretKeyBytes)
	config.SecretKey = string(secretKeyBytes)

	privateKeyPath := path.Join(keyPath, privateKeyFile)

	keyBytes, err := ioutil.ReadFile(privateKeyPath)
	if err != nil {
		return config, fmt.Errorf("unable to read private key path: %s, error: %s", privateKeyPath, err)
	}

	config.PrivateKey = string(keyBytes)

	if val, ok := os.LookupEnv("application_id"); ok && len(val) > 0 {
		config.ApplicationID = val
	} else {
		return config, fmt.Errorf("application_id must be given")
	}

	return config, nil
}

func getSecretPath() (string, error) {
	secretPath := os.Getenv("secret_path")

	if len(secretPath) == 0 {
		return "", fmt.Errorf("secret_path env-var not set, this should be /var/openfaas/secrets or /run/secrets")
	}

	return secretPath, nil
}

func getFirstLine(secret []byte) []byte {
	stringSecret := string(secret)
	if newLine := strings.Index(stringSecret, "\n"); newLine != -1 {
		secret = secret[:newLine]
	}
	return secret
}
