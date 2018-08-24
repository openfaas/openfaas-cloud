package sdk

import (
	"fmt"
	"os"

	"github.com/alexellis/hmac"
)

// HmacEnabled uses validate_hmac env-var to verify if the
// feature is enabled
func HmacEnabled() bool {
	return os.Getenv("validate_hmac") == "1" || os.Getenv("validate_hmac") == "true"
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
