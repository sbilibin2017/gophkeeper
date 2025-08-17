package aesgsm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAESGCM_EncryptDecrypt(t *testing.T) {
	aesgcm, err := New()
	assert.NoError(t, err)
	assert.Len(t, aesgcm.Key, 32)

	tests := []struct {
		name      string
		plaintext []byte
	}{
		{"empty", []byte{}},
		{"short text", []byte("hello world")},
		{"long text", []byte("this is a longer text to test AES GCM encryption and decryption")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ciphertext, nonce, err := aesgcm.Encrypt(tt.plaintext)
			assert.NoError(t, err)
			assert.NotNil(t, ciphertext)
			assert.NotNil(t, nonce)
			assert.NotEqual(t, tt.plaintext, ciphertext)

			decrypted, err := aesgcm.Decrypt(ciphertext, nonce)
			assert.NoError(t, err)
			assert.Equal(t, tt.plaintext, decrypted)
		})
	}
}

func TestAESGCM_NonceUniqueness(t *testing.T) {
	aesgcm, err := New()
	assert.NoError(t, err)

	data := []byte("same data multiple times")
	nonces := make(map[string]struct{})

	for i := 0; i < 10; i++ {
		_, nonce, err := aesgcm.Encrypt(data)
		assert.NoError(t, err)
		nonceStr := string(nonce)
		_, exists := nonces[nonceStr]
		assert.False(t, exists, "nonce should be unique each time")
		nonces[nonceStr] = struct{}{}
	}
}
