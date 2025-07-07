package models

import "time"

const (
	TypeLoginPassword = "login_password"
	TypeText          = "text"
	TypeBinary        = "binary"
	TypeCard          = "card"
)

// SecretLoginPassword represents login-password credentials with optional metadata.
type LoginPassword struct {
	SecretID  string            `json:"secret_id" db:"secret_id"`   // Unique identifier for the secret
	Login     string            `json:"login" db:"login"`           // The login or username
	Password  string            `json:"password" db:"password"`     // The associated password
	Meta      map[string]string `json:"meta,omitempty" db:"meta"`   // Optional metadata as key-value pairs
	UpdatedAt time.Time         `json:"updated_at" db:"updated_at"` // Timestamp of the last update
}

// SecretText represents a textual secret with optional metadata.
type Text struct {
	SecretID  string            `json:"secret_id" db:"secret_id"`   // Unique identifier for the secret
	Content   string            `json:"content" db:"content"`       // The main text content of the secret
	Meta      map[string]string `json:"meta,omitempty" db:"meta"`   // Optional metadata as key-value pairs
	UpdatedAt time.Time         `json:"updated_at" db:"updated_at"` // Timestamp of the last update
}

// SecretBinary represents a binary secret (e.g., file data) with optional metadata.
type Binary struct {
	SecretID  string            `json:"secret_id" db:"secret_id"`   // Unique identifier for the secret
	Data      []byte            `json:"data" db:"data"`             // Raw binary data
	Meta      map[string]string `json:"meta,omitempty" db:"meta"`   // Optional metadata as key-value pairs
	UpdatedAt time.Time         `json:"updated_at" db:"updated_at"` // Timestamp of the last update
}

// SecretCard represents sensitive card information with optional metadata.
type Card struct {
	SecretID  string            `json:"secret_id" db:"secret_id"`   // Unique identifier for the secret
	Number    string            `json:"number" db:"number"`         // Card number
	Holder    string            `json:"holder" db:"holder"`         // Name of the cardholder
	ExpMonth  int               `json:"exp_month" db:"exp_month"`   // Expiration month (1–12)
	ExpYear   int               `json:"exp_year" db:"exp_year"`     // Expiration year (4-digit)
	CVV       string            `json:"cvv" db:"cvv"`               // Card verification value (CVV)
	Meta      map[string]string `json:"meta,omitempty" db:"meta"`   // Optional metadata as key-value pairs
	UpdatedAt time.Time         `json:"updated_at" db:"updated_at"` // Timestamp of the last update
}
