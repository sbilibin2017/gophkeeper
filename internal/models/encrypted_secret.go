package models

// EncryptedSecret represents a hybrid-encrypted secret.
type EncryptedSecret struct {
	SecretName string // Identifier for the secret
	SecretType string // Type/category of the secret (e.g., bankcard, user, text)
	Ciphertext []byte // AES-GCM encrypted data, including nonce, ciphertext, and tag
	AESKeyEnc  []byte // AES key encrypted using RSA-OAEP
	Timestamp  int64  // Unix timestamp indicating when the secret was encrypted
}
