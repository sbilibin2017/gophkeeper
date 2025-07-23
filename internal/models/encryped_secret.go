package models

// EncrypedSecret represents a securely encoded secret using asymmetric encryption combined with HMAC for integrity verification.
type EncrypedSecret struct {
	SecretName string `json:"secret_name"`         // Name of the secret
	SecretType string `json:"secret_type"`         // Type of the secret (e.g., bankcard, user, text)
	Ciphertext []byte `json:"ciphertext"`          // Encrypted secret data (output of asymmetric encryption)
	HMAC       []byte `json:"hmac"`                // HMAC for verifying integrity/authenticity
	Nonce      []byte `json:"nonce,omitempty"`     // Optional nonce/IV for hybrid encryption (e.g., AES-GCM)
	KeyID      string `json:"key_id,omitempty"`    // Optional identifier of public key used for encryption
	Timestamp  int64  `json:"timestamp,omitempty"` // Optional Unix timestamp when encryption occurred
}
