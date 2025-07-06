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

type SecretDB struct {
	SecretID  string    `json:"secret_id" db:"secret_id"`   // Unique identifier of the secret (UUID)
	TypeID    string    `json:"type_id" db:"type_id"`       // Type identifier of the secret
	OwnerID   string    `json:"owner_id" db:"owner_id"`     // Identifier of the secret owner (UUID)
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"` // Date and time of last update
}

type TypeDB struct {
	TypeID    string    `json:"type_id" db:"type_id"`
	Name      string    `json:"name" db:"name"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type LoginPasswordDB struct {
	SecretID string            `json:"secret_id" db:"secret_id"` // Unique identifier of the secret (UUID)
	Login    string            `json:"login" db:"login"`
	Password string            `json:"password" db:"password"`
	Meta     map[string]string `json:"meta,omitempty" db:"meta"` // Assuming meta is stored as JSON in DB
}

// PayloadText describes the data of a secret of type "text".
type PayloadTextDB struct {
	SecretID string            `json:"secret_id" db:"secret_id"` // Unique identifier of the secret (UUID)
	Content  string            `json:"content" db:"content"`     // Main text content
	Meta     map[string]string `json:"meta,omitempty" db:"meta"` // Additional metadata
}

// PayloadBinary describes the data of a secret of type "binary data".
type PayloadBinaryDB struct {
	SecretID string            `json:"secret_id" db:"secret_id"` // Unique identifier of the secret (UUID)
	Data     []byte            `json:"data" db:"data"`           // Raw binary data
	Meta     map[string]string `json:"meta,omitempty" db:"meta"` // Additional metadata
}

// PayloadCard describes the data of a secret of type "payment card".
type PayloadCardDB struct {
	SecretID string            `json:"secret_id" db:"secret_id"` // Unique identifier of the secret (UUID)
	Number   string            `json:"number" db:"number"`       // Card number
	Holder   string            `json:"holder" db:"holder"`       // Cardholder name
	ExpMonth int               `json:"exp_month" db:"exp_month"` // Expiration month
	ExpYear  int               `json:"exp_year" db:"exp_year"`   // Expiration year
	CVV      string            `json:"cvv" db:"cvv"`             // CVV code
	Meta     map[string]string `json:"meta,omitempty" db:"meta"` // Additional metadata (e.g., bank, card type)
}
