package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
)

// EncryptedSecret represents a hybrid-encrypted secret.
type Encrypted struct {
	Ciphertext []byte // AES-GCM encrypted data, including nonce, ciphertext, and tag
	AESKeyEnc  []byte // AES key encrypted using RSA-OAEP
}

// Encryptor holds the RSA public key for encryption.
type Encryptor struct {
	pubKey *rsa.PublicKey
}

// NewEncryptor creates a new Encryptor by loading the RSA public key from a PEM cert file.
func NewEncryptor(certFilePath string) (*Encryptor, error) {
	certPEM, err := os.ReadFile(certFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read cert file: %w", err)
	}

	block, _ := pem.Decode(certPEM)
	if block == nil {
		return nil, fmt.Errorf("failed to parse PEM block from cert file")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate: %w", err)
	}

	pubKey, ok := cert.PublicKey.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("certificate does not contain an RSA public key")
	}

	return &Encryptor{pubKey: pubKey}, nil
}

// Encrypt encrypts the plaintext using AES-GCM and encrypts the AES key with RSA-OAEP.
func (e *Encryptor) Encrypt(plaintext []byte) (*Encrypted, error) {
	// Generate AES key
	aesKey := make([]byte, 32) // AES-256
	if _, err := rand.Read(aesKey); err != nil {
		return nil, fmt.Errorf("AES key gen failed: %w", err)
	}

	// Encrypt data using AES-GCM
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, fmt.Errorf("AES cipher init failed: %w", err)
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("AES-GCM init failed: %w", err)
	}
	nonce := make([]byte, aead.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, fmt.Errorf("nonce gen failed: %w", err)
	}
	ciphertext := aead.Seal(nonce, nonce, plaintext, nil) // nonce | ciphertext | tag

	// Encrypt AES key with RSA-OAEP
	label := []byte("AESKey")
	encryptedAESKey, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, e.pubKey, aesKey, label)
	if err != nil {
		return nil, fmt.Errorf("RSA encryption failed: %w", err)
	}

	return &Encrypted{
		Ciphertext: ciphertext,
		AESKeyEnc:  encryptedAESKey,
	}, nil
}

// Decryptor holds the RSA private key for decryption.
type Decryptor struct {
	privKey *rsa.PrivateKey
}

// NewDecryptor loads an RSA private key from a PEM file and returns a Decryptor.
func NewDecryptor(privKeyPath string) (*Decryptor, error) {
	privPEM, err := os.ReadFile(privKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read private key file: %w", err)
	}

	block, _ := pem.Decode(privPEM)
	if block == nil {
		return nil, fmt.Errorf("failed to parse PEM block from private key file")
	}

	privKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		key, err2 := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err2 != nil {
			return nil, fmt.Errorf("failed to parse private key: %v, %v", err, err2)
		}
		var ok bool
		privKey, ok = key.(*rsa.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("private key is not RSA")
		}
	}

	return &Decryptor{privKey: privKey}, nil
}

// Decrypt decrypts the encrypted data using RSA-OAEP and AES-GCM.
func (d *Decryptor) Decrypt(encrypted *Encrypted) ([]byte, error) {
	// 1. Decrypt AES key with RSA-OAEP
	label := []byte("AESKey")
	aesKey, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, d.privKey, encrypted.AESKeyEnc, label)
	if err != nil {
		return nil, fmt.Errorf("RSA decryption failed: %w", err)
	}

	// 2. Decrypt data using AES-GCM
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, fmt.Errorf("AES cipher init failed: %w", err)
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("AES-GCM init failed: %w", err)
	}

	nonceSize := aead.NonceSize()
	if len(encrypted.Ciphertext) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce := encrypted.Ciphertext[:nonceSize]
	ciphertext := encrypted.Ciphertext[nonceSize:]

	plaintext, err := aead.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("AES-GCM decryption failed: %w", err)
	}

	return plaintext, nil
}
