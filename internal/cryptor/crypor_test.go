package cryptor

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"testing"
	"time"

	"github.com/sbilibin2017/gophkeeper/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func generateRSAKeys(t *testing.T) (*rsa.PrivateKey, *rsa.PublicKey) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	return priv, &priv.PublicKey
}

func encodePrivateKeyPEM(priv *rsa.PrivateKey) []byte {
	return pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(priv),
		},
	)
}

func generateSelfSignedCertPEM(t *testing.T, priv *rsa.PrivateKey) []byte {
	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName: "Test Cert",
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	certDER, err := x509.CreateCertificate(rand.Reader, template, template, &priv.PublicKey, priv)
	require.NoError(t, err)

	return pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certDER,
	})
}

func TestEncryptDecrypt_Success(t *testing.T) {
	priv, _ := generateRSAKeys(t)

	certPEM := generateSelfSignedCertPEM(t, priv)
	privPEM := encodePrivateKeyPEM(priv)

	c, err := New(
		WithPrivateKeyPEM(privPEM),
		WithPublicKeyPEM(certPEM),
	)
	require.NoError(t, err)
	require.NotNil(t, c)

	plaintext := []byte("super secret data")

	enc, err := c.Encrypt(plaintext)
	require.NoError(t, err)
	require.NotNil(t, enc)
	assert.NotEmpty(t, enc.Ciphertext)
	assert.NotEmpty(t, enc.AESKeyEnc)

	dec, err := c.Decrypt(enc)
	require.NoError(t, err)
	assert.Equal(t, plaintext, dec)
}

func TestEncrypt_NoPublicKey(t *testing.T) {
	c := &Cryptor{}

	_, err := c.Encrypt([]byte("data"))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "public key is nil")
}

func TestDecrypt_NoPrivateKey(t *testing.T) {
	c := &Cryptor{}

	_, err := c.Decrypt(&models.SecretEncrypted{}) // fixed type name
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "private key is nil")
}

func TestDecrypt_InvalidAESKey(t *testing.T) {
	priv, _ := generateRSAKeys(t)

	certPEM := generateSelfSignedCertPEM(t, priv)
	privPEM := encodePrivateKeyPEM(priv)

	c, err := New(
		WithPrivateKeyPEM(privPEM),
		WithPublicKeyPEM(certPEM),
	)
	require.NoError(t, err)

	enc, err := c.Encrypt([]byte("data"))
	require.NoError(t, err)

	// Corrupt the AES key
	enc.AESKeyEnc[0] ^= 0xFF

	_, err = c.Decrypt(enc)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "RSA decryption failed")
}

func TestDecrypt_CorruptCiphertext(t *testing.T) {
	priv, _ := generateRSAKeys(t)

	certPEM := generateSelfSignedCertPEM(t, priv)
	privPEM := encodePrivateKeyPEM(priv)

	c, err := New(
		WithPrivateKeyPEM(privPEM),
		WithPublicKeyPEM(certPEM),
	)
	require.NoError(t, err)

	plaintext := []byte("hello world")
	enc, err := c.Encrypt(plaintext)
	require.NoError(t, err)

	// Corrupt the ciphertext
	enc.Ciphertext[5] ^= 0xFF

	_, err = c.Decrypt(enc)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "AES decryption failed")
}

func TestWithPrivateKeyPEM(t *testing.T) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	pkcs1PEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(priv),
	})

	pkcs8Bytes, err := x509.MarshalPKCS8PrivateKey(priv)
	require.NoError(t, err)
	pkcs8PEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: pkcs8Bytes,
	})

	invalidPEM := []byte("-----BEGIN PRIVATE KEY-----\ninvaliddata\n-----END PRIVATE KEY-----")
	notPEM := []byte("not a pem block")

	pubBytes, err := x509.MarshalPKIXPublicKey(&priv.PublicKey)
	require.NoError(t, err)
	pubPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubBytes,
	})

	tests := []struct {
		name        string
		pemBytes    []byte
		expectError bool
		errorText   string
	}{
		{
			name:        "Valid PKCS#1 private key PEM",
			pemBytes:    pkcs1PEM,
			expectError: false,
		},
		{
			name:        "Valid PKCS#8 private key PEM",
			pemBytes:    pkcs8PEM,
			expectError: false,
		},
		{
			name:        "Invalid PEM block",
			pemBytes:    invalidPEM,
			expectError: true,
			errorText:   "invalid private key PEM block",
		},
		{
			name:        "Not a PEM block",
			pemBytes:    notPEM,
			expectError: true,
			errorText:   "invalid private key PEM block",
		},
		{
			name:        "Non-RSA private key",
			pemBytes:    pubPEM,
			expectError: true,
			errorText:   "parse private key failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Cryptor{}
			err := WithPrivateKeyPEM(tt.pemBytes)(c)

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorText)
			} else {
				require.NoError(t, err)
				require.NotNil(t, c.PrivateKey)
			}
		})
	}
}
