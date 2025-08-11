package rsa

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
)

// GenerateRSAKeys generates a new RSA key pair and returns the PEM-encoded public and private keys.
func GenerateRSAKeys(username string) (pubPEM []byte, privPEM []byte, err error) {
	// Generate a 2048-bit private key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}

	// Encode private key to PKCS1 PEM format
	privPEM = pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
		},
	)

	// Marshal public key to PKIX ASN.1 DER encoded form
	pubASN1, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return nil, nil, err
	}

	// Encode public key to PEM format
	pubPEM = pem.EncodeToMemory(
		&pem.Block{
			Type:  "PUBLIC KEY",
			Bytes: pubASN1,
		},
	)

	return pubPEM, privPEM, nil
}

// RSAKeyPair holds the PEM-encoded RSA keys as strings for easy JSON marshalling.
type RSAKeyPair struct {
	PublicKey  string `json:"public_key"`
	PrivateKey string `json:"private_key"`
}

// SaveKeyPair saves the given RSA keys to a JSON file at ~/.config/{username}.json
func SaveKeyPair(username string, pubPEM, privPEM []byte) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ".config")
	if err := os.MkdirAll(configDir, 0o700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	filePath := filepath.Join(configDir, fmt.Sprintf("%s.json", username))

	keyPair := RSAKeyPair{
		PublicKey:  string(pubPEM),
		PrivateKey: string(privPEM),
	}

	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		return fmt.Errorf("failed to open file for writing: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(&keyPair); err != nil {
		return fmt.Errorf("failed to encode key pair to JSON: %w", err)
	}

	return nil
}

// GetKeyPair reads the RSA key pair JSON file at ~/.config/{username}.json
// and returns the public and private keys as byte slices.
func GetKeyPair(username string) (pubPEM []byte, privPEM []byte, err error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get user home directory: %w", err)
	}

	filePath := filepath.Join(homeDir, ".config", fmt.Sprintf("%s.json", username))

	file, err := os.Open(filePath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open key pair file: %w", err)
	}
	defer file.Close()

	var keyPair RSAKeyPair
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&keyPair); err != nil {
		return nil, nil, fmt.Errorf("failed to decode key pair JSON: %w", err)
	}

	return []byte(keyPair.PublicKey), []byte(keyPair.PrivateKey), nil
}
