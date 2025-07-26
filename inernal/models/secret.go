package models

import (
	"time"
)

const (
	SecretTypeBankCard = "bankcard" // Secret type: BankCard
	SecretTypeUser     = "user"     // Secret type: User
	SecretTypeText     = "text"     // Secret type: Text
	SecretTypeBinary   = "binary"   // Secret type: Binary
)

const (
	SyncModeServer      = "server"      // Sync handled by the server
	SyncModeClient      = "client"      // Sync handled by the client
	SyncModeInteractive = "interactive" // Sync done interactively with user input
)

// SecretSaveRequest represents a request to save a secret, including encryption data and user token.
type SecretSaveRequest struct {
	SecretName string `json:"secret_name"` // Secret identifier
	SecretType string `json:"secret_type"` // Secret type (e.g. bankcard, user)
	Ciphertext []byte `json:"ciphertext"`  // Encrypted secret data (AES-GCM)
	AESKeyEnc  []byte `json:"aes_key_enc"` // Encrypted AES key (RSA-OAEP)
	Token      string `json:"token"`       // User JWT token for authorization
}

// SecretGetRequest represents a request to get a secret by name, type, and token.
type SecretGetRequest struct {
	SecretName string `json:"secret_name"`
	SecretType string `json:"secret_type"`
	Token      string `json:"token"`
}

// SecretListRequest represents a request to list all secrets for a user token.
type SecretListRequest struct {
	Token string `json:"token"`
}

// SecretDB represents a secret stored in the database, including timestamps.
type SecretDB struct {
	SecretName  string    `json:"secret_name" db:"secret_name"`
	SecretType  string    `json:"secret_type" db:"secret_type"`
	SecretOwner string    `json:"secret_owner" db:"secret_owner"`
	Ciphertext  []byte    `json:"ciphertext" db:"ciphertext"`
	AESKeyEnc   []byte    `json:"aes_key_enc" db:"aes_key_enc"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type SecretGetFilterDB struct {
	SecretName  string `json:"secret_name"`
	SecretType  string `json:"secret_type"`
	SecretOwner string `json:"secret_owner"`
}

type SecretListFilterDB struct {
	SecretOwner string `json:"secret_owner"`
}

// BankcardSecret represents the decrypted bank card secret fields.
type BankcardSecretAdd struct {
	SecretName  string  `json:"secret_name"`
	SecretType  string  `json:"secret_type"`
	SecretOwner string  `json:"secret_owner"`
	Number      string  `json:"number"`
	Owner       string  `json:"owner"`
	Exp         string  `json:"exp"`
	CVV         string  `json:"cvv"`
	Meta        *string `json:"meta,omitempty"`
}

// BankcardSecret represents the decrypted bank card secret fields.
type BankcardSecretPayload struct {
	Number string  `json:"number"`
	Owner  string  `json:"owner"`
	Exp    string  `json:"exp"`
	CVV    string  `json:"cvv"`
	Meta   *string `json:"meta,omitempty"`
}

// UserSecretAdd represents the user secret fields along with metadata.
type UserSecretAdd struct {
	SecretName  string  `json:"secret_name"`
	SecretType  string  `json:"secret_type"`
	SecretOwner string  `json:"secret_owner"`
	Username    string  `json:"username"`
	Password    string  `json:"password"`
	Meta        *string `json:"meta,omitempty"`
}

// BinarySecretAdd represents the binary secret fields along with metadata.
type BinarySecretAdd struct {
	SecretName  string  `json:"secret_name"`
	SecretType  string  `json:"secret_type"`
	SecretOwner string  `json:"secret_owner"`
	Data        []byte  `json:"data"`
	Meta        *string `json:"meta,omitempty"`
}

// TextSecretAdd represents the text secret fields along with metadata.
type TextSecretAdd struct {
	SecretName  string  `json:"secret_name"`
	SecretType  string  `json:"secret_type"`
	SecretOwner string  `json:"secret_owner"`
	Text        string  `json:"text"`
	Meta        *string `json:"meta,omitempty"`
}

// UserSecretPayload represents the decrypted user secret fields.
type UserSecretPayload struct {
	Username string  `json:"username"`
	Password string  `json:"password"`
	Meta     *string `json:"meta,omitempty"`
}

// BinarySecretPayload represents the decrypted binary secret fields.
type BinarySecretPayload struct {
	Data []byte  `json:"data"`
	Meta *string `json:"meta,omitempty"`
}

// TextSecretPayload represents the decrypted text secret fields.
type TextSecretPayload struct {
	Text string  `json:"text"`
	Meta *string `json:"meta,omitempty"`
}
