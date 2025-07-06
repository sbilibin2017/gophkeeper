package models

import (
	"time"
)

// Constants defining types of stored secrets.
const (
	LoginPassword = "login_password" // Login and password
	Text          = "text"           // Text information
	Binary        = "binary"         // Binary data
	Card          = "card"           // Payment card
)

// SecretDB represents the main metadata for a stored secret.
// It contains identifiers and timestamps common for all secret types.
type SecretDB struct {
	SecretID  string    `json:"secret_id" db:"secret_id"`   // Unique identifier of the secret (UUID)
	TypeID    string    `json:"type_id" db:"type_id"`       // Type identifier of the secret, referencing TypeDB.TypeID
	OwnerID   string    `json:"owner_id" db:"owner_id"`     // Identifier of the secret owner (UUID)
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"` // Timestamp of last update
}

// TypeDB represents the type definition for secrets.
type TypeDB struct {
	TypeID    string    `json:"type_id" db:"type_id"`       // Unique identifier of the secret type
	Name      string    `json:"name" db:"name"`             // Human-readable name of the secret type
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"` // Timestamp of last update
}

// LoginPasswordDB represents a secret of type "login_password".
// It stores login credentials and optional metadata.
type LoginPasswordDB struct {
	SecretID string            `json:"secret_id" db:"secret_id"` // Unique identifier of the secret (UUID)
	Login    string            `json:"login" db:"login"`         // Username or login
	Password string            `json:"password" db:"password"`   // Password
	Meta     map[string]string `json:"meta,omitempty" db:"meta"` // Optional metadata
}

// PayloadTextDB represents a secret of type "text".
// It stores arbitrary text content and optional metadata.
type PayloadTextDB struct {
	SecretID string            `json:"secret_id" db:"secret_id"` // Unique identifier of the secret (UUID)
	Content  string            `json:"content" db:"content"`     // Main text content
	Meta     map[string]string `json:"meta,omitempty" db:"meta"` // Optional metadata
}

// PayloadBinaryDB represents a secret of type "binary".
// It stores raw binary data and optional metadata.
type PayloadBinaryDB struct {
	SecretID string            `json:"secret_id" db:"secret_id"` // Unique identifier of the secret (UUID)
	Data     []byte            `json:"data" db:"data"`           // Raw binary data
	Meta     map[string]string `json:"meta,omitempty" db:"meta"` // Optional metadata
}

// PayloadCardDB represents a secret of type "card".
// It stores payment card details and optional metadata.
type PayloadCardDB struct {
	SecretID string            `json:"secret_id" db:"secret_id"` // Unique identifier of the secret (UUID)
	Number   string            `json:"number" db:"number"`       // Card number
	Holder   string            `json:"holder" db:"holder"`       // Cardholder name
	ExpMonth int               `json:"exp_month" db:"exp_month"` // Expiration month (1-12)
	ExpYear  int               `json:"exp_year" db:"exp_year"`   // Expiration year (four digits)
	CVV      string            `json:"cvv" db:"cvv"`             // CVV code
	Meta     map[string]string `json:"meta,omitempty" db:"meta"` // Optional metadata (e.g., bank, card type)
}
