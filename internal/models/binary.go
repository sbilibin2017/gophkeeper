package models

import (
	"time"

	"github.com/sbilibin2017/gophkeeper/internal/models/fields"
)

// BinaryAddRequest represents the request payload for adding a new binary secret.
type BinaryAddRequest struct {
	SecretName string            `json:"secret_name,omitempty" validate:"required"` // Unique name of the secret
	Data       []byte            `json:"data,omitempty" validate:"required"`        // Binary data of the secret
	Meta       *fields.StringMap `json:"meta,omitempty"`                            // Additional metadata or notes (optional)
}

// BinaryFilterRequest represents the request to get a binary secret by its name.
type BinaryFilterRequest struct {
	SecretName string `json:"secret_name,omitempty"` // Unique name of the secret to retrieve
}

// BinaryDB represents a stored binary secret in the database.
type BinaryDB struct {
	SecretName  string            `json:"secret_name" db:"secret_name"`   // Unique name of the secret
	SecretOwner string            `json:"secret_owner" db:"secret_owner"` // Owner identifier of the secret
	Data        []byte            `json:"data" db:"data"`                 // Binary data of the secret
	Meta        *fields.StringMap `json:"meta,omitempty" db:"meta"`       // Additional metadata or notes (optional)
	UpdatedAt   time.Time         `json:"updated_at" db:"updated_at"`     // Last update timestamp in string format
}
