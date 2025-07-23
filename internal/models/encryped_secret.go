package models

// EncryptedSecret represents a hybrid-encrypted secret:
type EncryptedSecret struct {
	SecretName string `json:"secret_name"`         // Identifier for the secret
	SecretType string `json:"secret_type"`         // Type of secret (e.g., bankcard, user, text)
	Ciphertext []byte `json:"ciphertext"`          // AES-GCM encrypted data
	HMAC       []byte `json:"hmac"`                // HMAC-SHA256 over ciphertext
	AESKeyEnc  []byte `json:"aes_key_enc"`         // AES key encrypted using RSA-OAEP
	Nonce      []byte `json:"nonce,omitempty"`     // AES-GCM nonce
	Timestamp  int64  `json:"timestamp,omitempty"` // Unix timestamp of encryption
}
