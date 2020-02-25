package sdk

import (
	"os"
	"path/filepath"
	"testing"
)

func Test_GetPrivateKeyPath(t *testing.T) {
	privateKeyNameCopy := os.Getenv("private_key_filename")
	secretMountPathCopy := os.Getenv("secret_mount_path")

	defer func() {
		os.Setenv("private_key_filename", privateKeyNameCopy)
		os.Setenv("secret_mount_path", secretMountPathCopy)
	}()

	os.Setenv("private_key_filename", "")
	os.Setenv("secret_mount_path", "")

	expected := filepath.Join(defaultSecretMountPath, defaultPrivateKeyName)
	got := GetPrivateKeyPath()

	if got != expected {
		t.Errorf("Expected: %s, Got: %s", expected, got)
	}

	os.Setenv("private_key_filename", "id_rsa")
	os.Setenv("secret_mount_path", "/etc/foo/bar")

	pkn := os.Getenv("private_key_filename")
	smp := os.Getenv("secret_mount_path")

	expected = filepath.Join(smp, pkn)
	got = GetPrivateKeyPath()

	if got != expected {
		t.Errorf("Expected: %s, Got: %s", expected, got)
	}
}
