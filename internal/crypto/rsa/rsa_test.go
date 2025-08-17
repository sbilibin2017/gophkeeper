package rsa

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRSA_GenerateEncryptDecrypt(t *testing.T) {
	rsaObj := New()

	privPEM, pubPEM, err := rsaObj.GenerateKeyPair()
	assert.NoError(t, err)
	assert.NotEmpty(t, privPEM)
	assert.NotEmpty(t, pubPEM)

	tests := []struct {
		name string
		data []byte
	}{
		{"normal", []byte("Hello, world!")},
		{"empty", []byte("")},
		{"long", []byte("This is a longer test string to check RSA encryption and decryption functionality")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Шифруем
			ciphertext, err := rsaObj.Encrypt(tt.data)
			assert.NoError(t, err)
			assert.NotEmpty(t, ciphertext)

			// Дешифруем
			plaintext, err := rsaObj.Decrypt(ciphertext)
			assert.NoError(t, err)
			assert.Equal(t, tt.data, plaintext)
		})
	}
}

func TestRSA_ParseKeys(t *testing.T) {
	rsaObj := New()
	privPEM, pubPEM, err := rsaObj.GenerateKeyPair()
	assert.NoError(t, err)

	t.Run("ParsePublicKey", func(t *testing.T) {
		r := New()
		err := r.ParsePublicKey([]byte(pubPEM))
		assert.NoError(t, err)
		assert.NotNil(t, r.PublicKey)
	})

	t.Run("ParsePrivateKey", func(t *testing.T) {
		r := New()
		err := r.ParsePrivateKey([]byte(privPEM))
		assert.NoError(t, err)
		assert.NotNil(t, r.PrivateKey)
	})
}

func TestRSA_EncryptDecrypt_Errors(t *testing.T) {
	r := New()
	_, err := r.Encrypt([]byte("data"))
	assert.Error(t, err)
	assert.EqualError(t, err, "public key is nil")

	_, err = r.Decrypt([]byte("data"))
	assert.Error(t, err)
	assert.EqualError(t, err, "private key is nil")
}

func TestRSA_ParseKeys_InvalidPEM(t *testing.T) {
	r := New()

	t.Run("InvalidPublicKey", func(t *testing.T) {
		err := r.ParsePublicKey([]byte("invalid pem"))
		assert.Error(t, err)
	})

	t.Run("InvalidPrivateKey", func(t *testing.T) {
		err := r.ParsePrivateKey([]byte("invalid pem"))
		assert.Error(t, err)
	})
}

func TestRSA_SetPrivateKey(t *testing.T) {
	r := New()

	// Генерируем временный ключ для теста
	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	assert.NoError(t, err)

	r.SetPrivateKey(privKey)
	assert.NotNil(t, r.PrivateKey)
	assert.Equal(t, privKey, r.PrivateKey)
}

func TestRSA_SetPublicKey(t *testing.T) {
	r := New()

	// Генерируем временный ключ для теста
	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	assert.NoError(t, err)
	pubKey := &privKey.PublicKey

	r.SetPublicKey(pubKey)
	assert.NotNil(t, r.PublicKey)
	assert.Equal(t, pubKey, r.PublicKey)
}
