package crypto

import (
	"testing"

	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Вспомогательная функция для генерации RSA ключа для тестов
func GenerateRSAKey() (*rsa.PrivateKey, error) {
	return rsa.GenerateKey(rand.Reader, 2048)
}

// Вспомогательная функция для кодирования приватного ключа в PEM
func EncodePrivateKeyToPEM(priv *rsa.PrivateKey) []byte {
	privBytes := x509.MarshalPKCS1PrivateKey(priv)
	block := &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privBytes,
	}
	return pem.EncodeToMemory(block)
}

// Вспомогательная функция для кодирования публичного ключа в PEM
func EncodePublicKeyToPEM(pub *rsa.PublicKey) []byte {
	pubBytes, _ := x509.MarshalPKIXPublicKey(pub)
	block := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubBytes,
	}
	return pem.EncodeToMemory(block)
}

func TestAESGCM_EncryptDecrypt(t *testing.T) {
	aesGCM, err := NewAESGCM()
	require.NoError(t, err)
	require.NotNil(t, aesGCM)

	plaintext := []byte("super secret data")

	ciphertext, nonce, err := aesGCM.Encrypt(plaintext)
	require.NoError(t, err)
	require.NotNil(t, ciphertext)
	require.NotNil(t, nonce)

	decrypted, err := aesGCM.Decrypt(ciphertext, nonce)
	require.NoError(t, err)
	assert.Equal(t, plaintext, decrypted)
}

func TestRSAEncryptor_EncryptDecryptAESKey(t *testing.T) {
	privKey, err := GenerateRSAKey()
	require.NoError(t, err)
	pubKey := &privKey.PublicKey

	rsaEnc := &RSAEncryptor{
		PublicKey:  pubKey,
		PrivateKey: privKey,
	}

	aesKey := make([]byte, 32)
	_, err = rand.Read(aesKey)
	require.NoError(t, err)

	encryptedKey, err := rsaEnc.EncryptAESKey(aesKey)
	require.NoError(t, err)
	require.NotNil(t, encryptedKey)

	decryptedKey, err := rsaEnc.DecryptAESKey(encryptedKey)
	require.NoError(t, err)
	assert.Equal(t, aesKey, decryptedKey)
}

func TestParseRSAPublicPrivateKeyPEM(t *testing.T) {
	privKey, err := GenerateRSAKey()
	require.NoError(t, err)
	pubKey := &privKey.PublicKey

	privPEM := EncodePrivateKeyToPEM(privKey)
	pubPEM := EncodePublicKeyToPEM(pubKey)

	parsedPriv, err := ParseRSAPrivateKeyPEM(privPEM)
	require.NoError(t, err)
	assert.Equal(t, privKey.D, parsedPriv.D)

	parsedPub, err := ParseRSAPublicKeyPEM(pubPEM)
	require.NoError(t, err)
	assert.Equal(t, pubKey.N, parsedPub.N)
}

func TestGenerateID(t *testing.T) {
	id1 := GenerateID()
	id2 := GenerateID()
	assert.NotEmpty(t, id1)
	assert.NotEmpty(t, id2)
	assert.NotEqual(t, id1, id2)
}
