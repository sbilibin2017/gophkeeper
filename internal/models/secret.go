package models

import "time"

const (
	SyncModeClient      = "client"
	SyncModeInteractive = "interactive"
)

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

// SecretGetRequest holds SecretEncrypted secret data.
type SecretGetRequest struct {
	SecretName string `json:"secret_name"`
	SecretType string `json:"secret_type"`
}

// Secret represents the secret storage structure in the database.
type SecretDB struct {
	SecretName  string    `json:"secret_name" db:"secret_name"`
	SecretType  string    `json:"secret_type" db:"secret_type"`
	SecretOwner string    `json:"secret_owner" db:"secret_owner"`
	Ciphertext  []byte    `json:"ciphertext" db:"ciphertext"`
	AESKeyEnc   []byte    `json:"aes_key_enc" db:"aes_key_enc"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// SecretBankcard represents a bank card secret payload.
type SecretBankcard struct {
	Number string  `json:"number"`
	Owner  string  `json:"owner"`
	Exp    string  `json:"exp"`
	CVV    string  `json:"cvv"`
	Meta   *string `json:"meta,omitempty"`
}

// SecretText represents a text secret payload.
type SecretText struct {
	Data string  `json:"data"`
	Meta *string `json:"meta,omitempty"`
}

// SecretBinary represents a binary secret payload.
type SecretBinary struct {
	Data []byte  `json:"data"`
	Meta *string `json:"meta,omitempty"`
}

// SecretUser represents a user secret payload.
type SecretUser struct {
	Username string  `json:"username"`
	Password string  `json:"password"`
	Meta     *string `json:"meta,omitempty"`
}
