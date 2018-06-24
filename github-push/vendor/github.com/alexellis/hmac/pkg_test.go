package hmac

import "testing"

func Test_GenerateInvalid_GivesError(t *testing.T) {

	input := []byte("test")
	signature := "ab"
	secretKey := "key"
	err := Validate(input, signature, secretKey)
	if err == nil {
		t.Errorf("expected error when signature didn't have at least 5 characters in length")
		t.Fail()
	}

	wantErr := "invalid xHubSignature, should have at least 5 characters"
	if err.Error() != wantErr {
		t.Errorf("want: %s, got: %s", wantErr, err.Error())
		t.Fail()
	}
}
