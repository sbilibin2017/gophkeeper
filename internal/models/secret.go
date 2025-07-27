package models

import "time"

// Secret types
const (
	SecretTypeBankCard = "bankcard"
	SecretTypeUser     = "user"
	SecretTypeText     = "text"
	SecretTypeBinary   = "binary"
)

// SecretEncrypted represents the secret storage structure in the database.
type SecretEncrypted struct {
	Ciphertext []byte `json:"ciphertext" db:"ciphertext"`
	AESKeyEnc  []byte `json:"aes_key_enc" db:"aes_key_enc"`
}

// SecretSaveRequest holds SecretEncrypted secret data.
type SecretSaveRequest struct {
	SecretName string `json:"secret_name"`
	SecretType string `json:"secret_type"`
	Ciphertext []byte `json:"ciphertext"`
	AESKeyEnc  []byte `json:"aes_key_enc"`
}

// SecretGetRequest holds SecretEncrypted secret data.
type SecretGetRequest struct {
	SecretName string `json:"secret_name"`
	SecretType string `json:"secret_type"`
}

// Secret represents the secret storage structure in the database.
type Secret struct {
	SecretName  string    `json:"secret_name" db:"secret_name"`
	SecretType  string    `json:"secret_type" db:"secret_type"`
	SecretOwner string    `json:"secret_owner" db:"secret_owner"`
	Ciphertext  []byte    `json:"ciphertext" db:"ciphertext"`
	AESKeyEnc   []byte    `json:"aes_key_enc" db:"aes_key_enc"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// BankcardPayload represents a bank card secret payload.
type BankcardPayload struct {
	Number string  `json:"number"`
	Owner  string  `json:"owner"`
	Exp    string  `json:"exp"`
	CVV    string  `json:"cvv"`
	Meta   *string `json:"meta,omitempty"`
}

// TextPayload represents a text secret payload.
type TextPayload struct {
	Data string  `json:"data"`
	Meta *string `json:"meta,omitempty"`
}

// BinaryPayload represents a binary secret payload.
type BinaryPayload struct {
	Data []byte  `json:"data"`
	Meta *string `json:"meta,omitempty"`
}

// UserPayload represents a user secret payload.
type UserPayload struct {
	Username string  `json:"username"`
	Password string  `json:"password"`
	Meta     *string `json:"meta,omitempty"`
}
