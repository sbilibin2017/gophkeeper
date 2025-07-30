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

	"github.com/sbilibin2017/gophkeeper/internal/models"
)

type Cryptor struct {
	PublicKey  *rsa.PublicKey
	PrivateKey *rsa.PrivateKey
}

// Opt defines a functional option for configuring a Cryptor.
type Opt func(*Cryptor) error

// New constructs a Cryptor using functional options.
func New(opts ...Opt) (*Cryptor, error) {
	c := &Cryptor{}
	for _, opt := range opts {
		if err := opt(c); err != nil {
			return nil, err
		}
	}
	return c, nil
}

// WithPublicKeyPEM sets the public key from PEM-encoded certificate bytes.
func WithPublicKeyPEM(pemBytes []byte) Opt {
	return func(c *Cryptor) error {
		block, _ := pem.Decode(pemBytes)
		if block == nil {
			return fmt.Errorf("invalid public key PEM block")
		}
		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return fmt.Errorf("parse certificate failed: %w", err)
		}
		pub, ok := cert.PublicKey.(*rsa.PublicKey)
		if !ok {
			return fmt.Errorf("certificate does not contain RSA public key")
		}
		c.PublicKey = pub
		return nil
	}
}

// WithPrivateKeyPEM sets the private key from PEM-encoded key bytes (PKCS#1 or PKCS#8).
func WithPrivateKeyPEM(pemBytes []byte) Opt {
	return func(c *Cryptor) error {
		block, _ := pem.Decode(pemBytes)
		if block == nil {
			return fmt.Errorf("invalid private key PEM block")
		}
		priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
		if err == nil {
			c.PrivateKey = priv
			return nil
		}
		key, err2 := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err2 != nil {
			return fmt.Errorf("parse private key failed: %v, %v", err, err2)
		}
		rsaPriv, ok := key.(*rsa.PrivateKey)
		if !ok {
			return fmt.Errorf("not an RSA private key")
		}
		c.PrivateKey = rsaPriv
		return nil
	}
}

// Encrypt performs hybrid encryption using the RSA public key.
func (c *Cryptor) Encrypt(plaintext []byte) (*models.SecretEncrypted, error) {
	if c.PublicKey == nil {
		return nil, fmt.Errorf("public key is nil")
	}
	aesKey := make([]byte, 32)
	if _, err := rand.Read(aesKey); err != nil {
		return nil, fmt.Errorf("AES key gen failed: %w", err)
	}

	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, fmt.Errorf("AES cipher failed: %w", err)
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("GCM failed: %w", err)
	}

	nonce := make([]byte, aead.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, fmt.Errorf("nonce gen failed: %w", err)
	}

	ciphertext := aead.Seal(nonce, nonce, plaintext, nil)

	label := []byte("AESKey")
	encKey, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, c.PublicKey, aesKey, label)
	if err != nil {
		return nil, fmt.Errorf("RSA encryption failed: %w", err)
	}

	return &models.SecretEncrypted{
		Ciphertext: ciphertext,
		AESKeyEnc:  encKey,
	}, nil
}

// Decrypt performs hybrid decryption using the RSA private key.
func (c *Cryptor) Decrypt(enc *models.SecretEncrypted) ([]byte, error) {
	if c.PrivateKey == nil {
		return nil, fmt.Errorf("private key is nil")
	}

	label := []byte("AESKey")
	aesKey, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, c.PrivateKey, enc.AESKeyEnc, label)
	if err != nil {
		return nil, fmt.Errorf("RSA decryption failed: %w", err)
	}

	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, fmt.Errorf("AES cipher failed: %w", err)
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("GCM failed: %w", err)
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
