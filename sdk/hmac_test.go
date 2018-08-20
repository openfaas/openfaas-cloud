package sdk

import (
	"encoding/hex"
	"fmt"
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
