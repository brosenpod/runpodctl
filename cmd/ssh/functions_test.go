package ssh

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/crypto/ssh"
)

func TestGenerateSSHKeyPair(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	keyName := "test-key"
	publicKey, err := GenerateSSHKeyPair(keyName)
	if err != nil {
		t.Fatalf("GenerateSSHKeyPair returned error: %v", err)
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("os.UserHomeDir returned error: %v", err)
	}

	privateKeyPath := filepath.Join(homeDir, ".runpod", "ssh", keyName)
	if _, err := os.Stat(privateKeyPath); err != nil {
		t.Fatalf("expected private key file %s to exist: %v", privateKeyPath, err)
	}

	publicKeyPath := privateKeyPath + ".pub"
	publicKeyBytes, err := os.ReadFile(publicKeyPath)
	if err != nil {
		t.Fatalf("reading public key file: %v", err)
	}

	if string(publicKeyBytes) != string(publicKey) {
		t.Fatalf("public key file contents differ from returned bytes")
	}

	expectedSuffix := " " + keyName + "\n"
	if !strings.HasSuffix(string(publicKey), expectedSuffix) {
		t.Fatalf("public key does not end with expected comment suffix %q: %q", expectedSuffix, publicKey)
	}
}

func TestGetLocalSSHKey(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	key, err := GetLocalSSHKey()
	if err != nil {
		t.Fatalf("GetLocalSSHKey returned error: %v", err)
	}
	if key != nil {
		t.Fatalf("expected no key to be returned when none exists")
	}

	generatedKey, err := GenerateSSHKeyPair("RunPod-Key-Go")
	if err != nil {
		t.Fatalf("GenerateSSHKeyPair returned error: %v", err)
	}

	key, err = GetLocalSSHKey()
	if err != nil {
		t.Fatalf("GetLocalSSHKey returned error after generation: %v", err)
	}
	if key == nil {
		t.Fatalf("expected key to be returned after generation")
	}

	if string(key) != string(generatedKey) {
		t.Fatalf("retrieved key does not match generated key")
	}

	if _, _, _, _, err := ssh.ParseAuthorizedKey(key); err != nil {
		t.Fatalf("ParseAuthorizedKey returned error: %v", err)
	}
}
