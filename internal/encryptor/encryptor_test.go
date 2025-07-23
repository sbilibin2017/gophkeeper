package encryptor

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"
	"time"

	"github.com/sbilibin2017/gophkeeper/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func generateRSAKeyPair(t *testing.T) *rsa.PrivateKey {
	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	return privKey
}

func TestEncryptDecrypt_Text(t *testing.T) {
	privKey := generateRSAKeyPair(t)
	pubKey := &privKey.PublicKey

	original := models.Text{
		SecretName:  "note-1",
		SecretOwner: "user123",
		Data:        "This is a secret note.",
		Meta:        strPtr("important"),
		UpdatedAt:   testTime(),
	}

	enc, err := Encrypt(pubKey, original.SecretName, "text", original)
	require.NoError(t, err)
	require.NotNil(t, enc)

	dec, err := Decrypt[models.Text](privKey, enc)
	require.NoError(t, err)
	require.Equal(t, original.SecretName, dec.SecretName)
	require.Equal(t, original.SecretOwner, dec.SecretOwner)
	require.Equal(t, original.Data, dec.Data)
	require.Equal(t, original.Meta, dec.Meta)
}

func TestEncryptDecrypt_User(t *testing.T) {
	privKey := generateRSAKeyPair(t)
	pubKey := &privKey.PublicKey

	original := models.User{
		SecretName:  "login-1",
		SecretOwner: "admin",
		Login:       "admin@example.com",
		Password:    "P@ssw0rd!",
		Meta:        strPtr("admin-account"),
		UpdatedAt:   testTime(),
	}

	enc, err := Encrypt(pubKey, original.SecretName, "user", original)
	require.NoError(t, err)
	require.NotNil(t, enc)

	dec, err := Decrypt[models.User](privKey, enc)
	require.NoError(t, err)
	require.Equal(t, original.Login, dec.Login)
	require.Equal(t, original.Password, dec.Password)
	require.Equal(t, original.Meta, dec.Meta)
}

func TestEncryptDecrypt_CustomStruct(t *testing.T) {
	privKey := generateRSAKeyPair(t)
	pubKey := &privKey.PublicKey

	type mockSecret struct {
		Title string
		Value string
	}
	original := mockSecret{
		Title: "API_KEY",
		Value: "12345-abcde",
	}

	enc, err := Encrypt(pubKey, "mock-1", "mock", original)
	require.NoError(t, err)

	dec, err := Decrypt[mockSecret](privKey, enc)
	require.NoError(t, err)
	require.Equal(t, original, *dec)
}

// --- Helper functions ---

func strPtr(s string) *string {
	return &s
}

func testTime() time.Time {
	return time.Now().UTC()
}

// Helper test struct
type testData struct {
	Field1 string
	Field2 int
}

// Generate test RSA key pair
func generateKeyPair(t *testing.T) (*rsa.PrivateKey, *rsa.PublicKey) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	assert.NoError(t, err)
	return priv, &priv.PublicKey
}

func TestEncryptDecrypt_Success(t *testing.T) {
	priv, pub := generateKeyPair(t)
	data := testData{"hello", 123}

	encrypted, err := Encrypt(pub, "mysecret", "test", data)
	assert.NoError(t, err)
	assert.NotNil(t, encrypted)

	decrypted, err := Decrypt[testData](priv, encrypted)
	assert.NoError(t, err)
	assert.NotNil(t, decrypted)
	assert.Equal(t, data, *decrypted)
}

func TestDecrypt_WrongKey(t *testing.T) {
	_, pub1 := generateKeyPair(t)
	priv2, _ := generateKeyPair(t)
	data := testData{"secret", 42}

	encrypted, err := Encrypt(pub1, "mysecret", "test", data)
	assert.NoError(t, err)

	// Try to decrypt with wrong private key
	decrypted, err := Decrypt[testData](priv2, encrypted)
	assert.Error(t, err)
	assert.Nil(t, decrypted)
}

func TestDecrypt_TamperedCiphertext(t *testing.T) {
	priv, pub := generateKeyPair(t)
	data := testData{"field", 99}

	encrypted, err := Encrypt(pub, "secret", "test", data)
	assert.NoError(t, err)

	// Tamper ciphertext
	if len(encrypted.Ciphertext) > 0 {
		encrypted.Ciphertext[0] ^= 0xFF
	}

	decrypted, err := Decrypt[testData](priv, encrypted)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "HMAC verification failed")
	assert.Nil(t, decrypted)
}

func TestDecrypt_CorruptedAESKey(t *testing.T) {
	priv, pub := generateKeyPair(t)
	data := testData{"data", 1}

	encrypted, err := Encrypt(pub, "secret", "test", data)
	assert.NoError(t, err)

	// Corrupt AES key ciphertext
	if len(encrypted.AESKeyEnc) > 0 {
		encrypted.AESKeyEnc[0] ^= 0xFF
	}

	decrypted, err := Decrypt[testData](priv, encrypted)
	assert.Error(t, err)
	assert.Nil(t, decrypted)
}

func TestEncrypt_InvalidData(t *testing.T) {
	pub := &rsa.PublicKey{} // invalid public key to force error

	// channel cannot be JSON marshaled
	ch := make(chan int)

	encrypted, err := Encrypt(pub, "secret", "test", ch)
	assert.Error(t, err)
	assert.Nil(t, encrypted)
}
