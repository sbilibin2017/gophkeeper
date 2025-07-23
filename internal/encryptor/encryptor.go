package encryptor

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"time"

	"github.com/sbilibin2017/gophkeeper/internal/models"
)

// Encrypt performs hybrid encryption:
// - AES-GCM for fast symmetric encryption
// - RSA-OAEP to encrypt the AES key
// - HMAC-SHA256 to ensure ciphertext integrity
func Encrypt[T any](pubKey *rsa.PublicKey, secretName, secretType string, data T) (*models.EncryptedSecret, error) {
	// Marshal input data to JSON
	plainData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	// Generate random AES-256 key
	aesKey := make([]byte, 32)
	if _, err := rand.Read(aesKey); err != nil {
		return nil, err
	}

	// AES-GCM block cipher
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// Generate nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}

	// Encrypt data using AES-GCM
	ciphertext := gcm.Seal(nil, nonce, plainData, nil)

	// Compute HMAC-SHA256 of ciphertext
	h := hmac.New(sha256.New, aesKey)
	h.Write(ciphertext)
	mac := h.Sum(nil)

	// Encrypt AES key using RSA-OAEP
	aesKeyEnc, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, pubKey, aesKey, nil)
	if err != nil {
		return nil, err
	}

	// Build and return EncryptedSecret
	return &models.EncryptedSecret{
		SecretName: secretName,
		SecretType: secretType,
		Ciphertext: ciphertext,
		HMAC:       mac,
		AESKeyEnc:  aesKeyEnc,
		Nonce:      nonce,
		Timestamp:  time.Now().Unix(),
	}, nil
}

// Decrypt performs hybrid decryption:
// - Decrypts the AES key using RSA-OAEP
// - Verifies HMAC-SHA256
// - Decrypts data using AES-GCM
func Decrypt[T any](privKey *rsa.PrivateKey, encrypted *models.EncryptedSecret) (*T, error) {
	// Decrypt AES key using RSA-OAEP
	aesKey, err := rsa.DecryptOAEP(sha256.New(), nil, privKey, encrypted.AESKeyEnc, nil)
	if err != nil {
		return nil, err
	}

	// Verify HMAC-SHA256
	h := hmac.New(sha256.New, aesKey)
	h.Write(encrypted.Ciphertext)
	expectedMAC := h.Sum(nil)

	if !hmac.Equal(expectedMAC, encrypted.HMAC) {
		return nil, errors.New("HMAC verification failed: data may be tampered")
	}

	// Decrypt data using AES-GCM
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	plaintext, err := gcm.Open(nil, encrypted.Nonce, encrypted.Ciphertext, nil)
	if err != nil {
		return nil, err
	}

	// Unmarshal into original type
	var result T
	if err := json.Unmarshal(plaintext, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
