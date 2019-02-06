package sdk

import (
	"encoding/hex"
	"fmt"
	"os"
	"testing"

	"github.com/alexellis/hmac"
)

func Test_validHMACWithSecretKey_validSecret(t *testing.T) {

	data := []byte("Store this string")
	key := []byte("key-goes-here")
	signed := hmac.Sign(data, key)
	digest := fmt.Sprintf("sha1=%s", hex.EncodeToString(signed))
	err := validHMACWithSecretKey(&data, string(key), digest)

	if err != nil {
		t.Errorf("with %s, found error: %s", digest, err)
		t.Fail()
	}
}

func Test_validHMACWithSecretKey_invalidSecret(t *testing.T) {

	data := []byte("Store this string")
	key := []byte("key-goes-here")
	signed := hmac.Sign(data, key)
	digest := fmt.Sprintf("sha1=%s", hex.EncodeToString(signed))
	err := validHMACWithSecretKey(&data, string(key[:4]), digest)

	if err == nil {
		t.Errorf("with %s, expected to find error", digest)
		t.Fail()
	}
}

func Test_HmacEnabled(t *testing.T) {
	tests := []struct {
		title        string
		value        string
		expectedBool bool
	}{
		{
			title:        "environmental variable `validate_hmac` is unset",
			value:        "",
			expectedBool: true,
		},
		{
			title:        "environmental variable `validate_hmac` is set with random value",
			value:        "random",
			expectedBool: true,
		},
		{
			title:        "environmental variable `validate_hmac` is set with explicit `0`",
			value:        "0",
			expectedBool: false,
		},
		{
			title:        "environmental variable `validate_hmac` is set with explicit `false`",
			value:        "false",
			expectedBool: false,
		},
	}
	hmacEnvVar := "validate_hmac"
	for _, test := range tests {
		t.Run(test.title, func(t *testing.T) {
			os.Setenv(hmacEnvVar, test.value)
			value := HmacEnabled()
			if value != test.expectedBool {
				t.Errorf("Expected value: %v got: %v", test.expectedBool, value)
			}
		})
	}
}
