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

// SecretDB represents the basic metadata of a stored secret.
// Contains identifiers and timestamps common to all secret types.
type SecretDB struct {
	SecretID  string    `json:"secret_id" db:"secret_id"`   // Unique secret identifier (UUID)
	TypeID    string    `json:"type_id" db:"type_id"`       // Secret type identifier, refers to TypeDB.TypeID
	OwnerID   string    `json:"owner_id" db:"owner_id"`     // Secret owner's identifier (UUID)
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"` // Timestamp of the last update
}

// TypeDB represents the definition of a secret type.
type TypeDB struct {
	TypeID    string    `json:"type_id" db:"type_id"`       // Unique identifier of the secret type
	Name      string    `json:"name" db:"name"`             // Human-readable name of the secret type
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"` // Timestamp of the last update
}

// LoginPasswordDB represents a secret of type "login_password".
// Stores credentials and optional metadata.
type LoginPasswordDB struct {
	SecretID string            `json:"secret_id" db:"secret_id"` // Unique secret identifier (UUID)
	Login    string            `json:"login" db:"login"`         // Username or login
	Password string            `json:"password" db:"password"`   // Password
	Meta     map[string]string `json:"meta,omitempty" db:"meta"` // Optional metadata
}

// PayloadTextDB represents a secret of type "text".
// Stores arbitrary text and optional metadata.
type PayloadTextDB struct {
	SecretID string            `json:"secret_id" db:"secret_id"` // Unique secret identifier (UUID)
	Content  string            `json:"content" db:"content"`     // Main text content
	Meta     map[string]string `json:"meta,omitempty" db:"meta"` // Optional metadata
}

// PayloadBinaryDB represents a secret of type "binary".
// Stores raw binary data and optional metadata.
type PayloadBinaryDB struct {
	SecretID string            `json:"secret_id" db:"secret_id"` // Unique secret identifier (UUID)
	Data     []byte            `json:"data" db:"data"`           // Raw binary data
	Meta     map[string]string `json:"meta,omitempty" db:"meta"` // Optional metadata
}

// PayloadCardDB represents a secret of type "card".
// Stores payment card data and optional metadata.
type PayloadCardDB struct {
	SecretID string            `json:"secret_id" db:"secret_id"` // Unique secret identifier (UUID)
	Number   string            `json:"number" db:"number"`       // Card number
	Holder   string            `json:"holder" db:"holder"`       // Cardholder's name
	ExpMonth int               `json:"exp_month" db:"exp_month"` // Expiry month (1-12)
	ExpYear  int               `json:"exp_year" db:"exp_year"`   // Expiry year (four digits)
	CVV      string            `json:"cvv" db:"cvv"`             // CVV code
	Meta     map[string]string `json:"meta,omitempty" db:"meta"` // Optional metadata (e.g., bank, card type)
}
