package models

// EncryptedSecret represents a hybrid-encrypted secret.
type EncryptedSecret struct {
	SecretName string `json:"secret_name" db:"secret_name"` // Identifier for the secret
	SecretType string `json:"secret_type" db:"secret_type"` // Type/category of the secret (e.g., bankcard, user, text)
	Ciphertext []byte `json:"ciphertext" db:"ciphertext"`   // AES-GCM encrypted data, including nonce, ciphertext, and tag
	AESKeyEnc  []byte `json:"aes_key_enc" db:"aes_key_enc"` // AES key encrypted using RSA-OAEP
	Timestamp  int64  `json:"timestamp" db:"timestamp"`     // Unix timestamp indicating when the secret was encrypted
}
