package cryptor

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

// Encrypted holds hybrid-encrypted data.
type Encrypted struct {
	Ciphertext []byte
	AESKeyEnc  []byte
}

// Cryptor supports hybrid encryption using RSA and AES-GCM.
type Cryptor struct {
	publicKey  *rsa.PublicKey
	privateKey *rsa.PrivateKey
}

// Opt defines a functional option for configuring HybridCrypto.
type Opt func(*Cryptor) error

// NewCryptor creates a new HybridCrypto using the provided functional options.
func New(opts ...Opt) (*Cryptor, error) {
	h := &Cryptor{}
	for _, opt := range opts {
		if err := opt(h); err != nil {
			return nil, err
		}
	}
	return h, nil
}

// WithPublicKeyFromCert loads a public key from an X.509 PEM certificate file.
func WithPublicKeyFromCert(certPath string) Opt {
	return func(h *Cryptor) error {
		data, err := os.ReadFile(certPath)
		if err != nil {
			return fmt.Errorf("failed to read cert file: %w", err)
		}
		block, _ := pem.Decode(data)
		if block == nil {
			return fmt.Errorf("invalid PEM block in cert")
		}
		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return fmt.Errorf("failed to parse certificate: %w", err)
		}
		pubKey, ok := cert.PublicKey.(*rsa.PublicKey)
		if !ok {
			return fmt.Errorf("certificate does not contain RSA public key")
		}
		h.publicKey = pubKey
		return nil
	}
}

// WithPrivateKeyFromFile loads a private key (PKCS#1 or PKCS#8) from a PEM file.
func WithPrivateKeyFromFile(keyPath string) Opt {
	return func(h *Cryptor) error {
		data, err := os.ReadFile(keyPath)
		if err != nil {
			return fmt.Errorf("failed to read private key file: %w", err)
		}
		block, _ := pem.Decode(data)
		if block == nil {
			return fmt.Errorf("invalid PEM block in private key")
		}

		// Try PKCS#1 first
		privKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
		if err == nil {
			h.privateKey = privKey
			return nil
		}

		// Try PKCS#8
		key, err2 := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err2 != nil {
			return fmt.Errorf("failed to parse private key: %v, %v", err, err2)
		}
		privKey, ok := key.(*rsa.PrivateKey)
		if !ok {
			return fmt.Errorf("not an RSA private key")
		}
		h.privateKey = privKey
		return nil
	}
}

// Encrypt performs hybrid encryption.
func (h *Cryptor) Encrypt(plaintext []byte) (*Encrypted, error) {
	if h.publicKey == nil {
		return nil, fmt.Errorf("public key is not set")
	}

	// Generate AES key
	aesKey := make([]byte, 32)
	if _, err := rand.Read(aesKey); err != nil {
		return nil, fmt.Errorf("AES key generation failed: %w", err)
	}

	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, fmt.Errorf("AES cipher creation failed: %w", err)
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("GCM creation failed: %w", err)
	}

	nonce := make([]byte, aead.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, fmt.Errorf("nonce generation failed: %w", err)
	}

	ciphertext := aead.Seal(nonce, nonce, plaintext, nil)

	label := []byte("AESKey")
	encKey, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, h.publicKey, aesKey, label)
	if err != nil {
		return nil, fmt.Errorf("RSA encryption failed: %w", err)
	}

	return &Encrypted{
		Ciphertext: ciphertext,
		AESKeyEnc:  encKey,
	}, nil
}

// Decrypt performs hybrid decryption.
func (h *Cryptor) Decrypt(enc *Encrypted) ([]byte, error) {
	if h.privateKey == nil {
		return nil, fmt.Errorf("private key is not set")
	}

	label := []byte("AESKey")
	aesKey, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, h.privateKey, enc.AESKeyEnc, label)
	if err != nil {
		return nil, fmt.Errorf("RSA decryption failed: %w", err)
	}

	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, fmt.Errorf("AES cipher creation failed: %w", err)
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("GCM creation failed: %w", err)
	}

	nonceSize := aead.NonceSize()
	if len(enc.Ciphertext) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}
	nonce := enc.Ciphertext[:nonceSize]
	ciphertext := enc.Ciphertext[nonceSize:]

	plaintext, err := aead.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("AES decryption failed: %w", err)
	}

	return plaintext, nil
}
