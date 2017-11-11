package function

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
)

// checkMAC verifies hash checksum
func checkMAC(message, messageMAC, key []byte) bool {
	mac := hmac.New(sha1.New, key)
	mac.Write(message)
	expectedMAC := mac.Sum(nil)

	return hmac.Equal(messageMAC, expectedMAC)
}

// validateHMAC validate a digest from Github via xHubSignature
func validateHMAC(bytesIn []byte, xHubSignature string, secretKey string) error {
	var validated error

	if len(xHubSignature) > 5 {
		messageMAC := xHubSignature[5:] // first few chars are: sha1=
		messageMACBuf, _ := hex.DecodeString(messageMAC)

		res := checkMAC(bytesIn, []byte(messageMACBuf), []byte(secretKey))
		if res == false {
			validated = fmt.Errorf("invalid message digest or secret")
		}
	}

	return validated
}
