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

// Secret represents a generic structure for storing a secret of any type.
type Secret struct {
	SecretID  string    `json:"secret_id" db:"secret_id"`   // Unique identifier of the secret (UUID)
	OwnerID   string    `json:"owner_id" db:"owner_id"`     // Identifier of the secret owner (UUID)
	SType     string    `json:"type" db:"type"`             // Type of secret (one of the constants above)
	Payload   []byte    `json:"payload" db:"payload"`       // Secret content in serialized form (JSON)
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"` // Date and time of last update
}

// PayloadLoginPassword describes the data of a secret of type "login and password".
type PayloadLoginPassword struct {
	Login    string            `json:"login"`          // Username or login
	Password string            `json:"password"`       // Password
	Meta     map[string]string `json:"meta,omitempty"` // Additional metadata (e.g., website, description, etc.)
}

// PayloadText describes the data of a secret of type "text".
type PayloadText struct {
	Content string            `json:"content"`        // Main text content
	Meta    map[string]string `json:"meta,omitempty"` // Additional metadata
}

// PayloadBinary describes the data of a secret of type "binary data".
type PayloadBinary struct {
	Data []byte            `json:"data"`           // Raw binary data
	Meta map[string]string `json:"meta,omitempty"` // Additional metadata
}

// PayloadCard describes the data of a secret of type "payment card".
type PayloadCard struct {
	Number   string            `json:"number"`         // Card number
	Holder   string            `json:"holder"`         // Cardholder name
	ExpMonth int               `json:"exp_month"`      // Expiration month
	ExpYear  int               `json:"exp_year"`       // Expiration year
	CVV      string            `json:"cvv"`            // CVV code
	Meta     map[string]string `json:"meta,omitempty"` // Additional metadata (e.g., bank, card type)
}
