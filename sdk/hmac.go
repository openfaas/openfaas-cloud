package sdk

import (
	"fmt"
	"os"

	"github.com/alexellis/hmac"
)

// HmacEnabled uses validate_hmac env-var to verify if the
// feature is disabled
func HmacEnabled() bool {
	if val, exists := os.LookupEnv("validate_hmac"); exists {
		return val != "false" && val != "0"
	}
	return true
}

// ValidHMAC returns an error if HMAC could not be validated or if
// the signature could not be loaded.
func ValidHMAC(payload *[]byte, secretKey string, digest string) error {
	key, err := ReadSecret(secretKey)
	if err != nil {
		return fmt.Errorf("unable to load HMAC symmetric key, %s", err.Error())
	}

	return validHMACWithSecretKey(payload, key, digest)
}

func validHMACWithSecretKey(payload *[]byte, secretText string, digest string) error {
	validated := hmac.Validate(*payload, digest, secretText)

	if validated != nil {
		return fmt.Errorf("unable to validate HMAC")
	}
	return nil
}

func readBool(key string) bool {
	if val, exists := os.LookupEnv(key); exists {
		return val != "false" && val != "0"
	}
	return true
}
