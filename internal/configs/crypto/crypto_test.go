package crypto

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"testing"
)

// TestHashPassword tests password hashing.
func TestHashPassword(t *testing.T) {
	password := "secret123"
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}
	if len(hash) == 0 {
		t.Fatal("HashPassword returned empty hash")
	}
}

// TestGenerateRSAKeys tests RSA key generation.
func TestGenerateRSAKeys(t *testing.T) {
	privKey, err := GenerateRSAKeys(2048)
	if err != nil {
		t.Fatalf("GenerateRSAKeys returned error: %v", err)
	}
	if privKey == nil {
		t.Fatal("GenerateRSAKeys returned nil private key")
	}
	if &privKey.PublicKey == nil {
		t.Fatal("GenerateRSAKeys returned nil public key")
	}
}

// TestGenerateDEK tests DEK generation.
func TestGenerateDEK(t *testing.T) {
	size := 32
	dek, err := GenerateDEK(size)
	if err != nil {
		t.Fatalf("GenerateDEK returned error: %v", err)
	}
	if len(dek) != size {
		t.Fatalf("GenerateDEK returned wrong size: got %d, want %d", len(dek), size)
	}
}

// TestEncryptDEK tests DEK encryption.
func TestEncryptDEK(t *testing.T) {
	privKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	pubKey := &privKey.PublicKey
	dek := []byte("mysecretdata")

	encrypted, err := EncryptDEK(pubKey, dek)
	if err != nil {
		t.Fatalf("EncryptDEK returned error: %v", err)
	}
	if len(encrypted) == 0 {
		t.Fatal("EncryptDEK returned empty ciphertext")
	}
}

// TestPrivateKeyToPEM tests PEM encoding of private key.
func TestRSAPrivateKeyToPEM(t *testing.T) {
	privKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	pemBytes := RSAPrivateKeyToPEM(privKey)

	if !bytes.HasPrefix(pemBytes, []byte("-----BEGIN RSA PRIVATE KEY-----")) {
		t.Fatalf("PEM encoding does not start with proper header")
	}
	if !bytes.HasSuffix(pemBytes, []byte("-----END RSA PRIVATE KEY-----\n")) {
		t.Fatalf("PEM encoding does not end with proper footer")
	}
}
