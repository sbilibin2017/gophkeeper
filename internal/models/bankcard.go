package models

import (
	"time"

	"github.com/sbilibin2017/gophkeeper/internal/models/fields"
)

// BankCardAddRequest represents the request payload for adding a new bank card secret.
type BankCardAddRequest struct {
	SecretName string            `json:"secret_name,omitempty" validate:"required"`       // Unique name of the secret
	Number     string            `json:"number,omitempty" validate:"required,luhn"`       // Card number (Luhn validated)
	Owner      string            `json:"owner,omitempty" validate:"required"`             // Card owner name
	Exp        string            `json:"exp,omitempty" validate:"required,len=5"`         // Expiration date (e.g., MM/YY)
	CVV        string            `json:"cvv,omitempty" validate:"required,len=3,numeric"` // Card CVV code
	Meta       *fields.StringMap `json:"meta,omitempty"`                                  // Additional metadata or notes
}

// BankCardFilterRequest represents the request to get a bank card secret by its name.
type BankCardFilterRequest struct {
	SecretName string `json:"secret_name,omitempty"` // Unique name of the secret to retrieve
}

// BankCardDB represents the stored bank card data including metadata and update time.
type BankCardDB struct {
	SecretName  string            `json:"secret_name,omitempty" db:"secret_name"`   // Unique name of the secret
	SecretOwner string            `json:"secret_owner,omitempty" db:"secret_owner"` // Owner of the secret (usually user or client ID)
	Number      string            `json:"number,omitempty" db:"number"`             // Card number
	Owner       string            `json:"owner,omitempty" db:"owner"`               // Card owner name
	Exp         string            `json:"exp,omitempty" db:"exp"`                   // Expiration date (e.g., MM/YY)
	CVV         string            `json:"cvv,omitempty" db:"cvv"`                   // Card CVV code
	Meta        *fields.StringMap `json:"meta,omitempty" db:"meta"`                 // Additional metadata or notes
	UpdatedAt   time.Time         `json:"updated_at,omitempty" db:"updated_at"`     // ISO8601 timestamp of last update
}
