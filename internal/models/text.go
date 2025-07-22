package models

import (
	"time"

	"github.com/sbilibin2017/gophkeeper/internal/models/fields"
)

// TextAddRequest represents the request payload for adding a new text secret.
type TextAddRequest struct {
	SecretName string            `json:"secret_name,omitempty" validate:"required"` // Unique name of the secret
	Content    string            `json:"content,omitempty" validate:"required"`     // Text content of the secret
	Meta       *fields.StringMap `json:"meta,omitempty"`                            // Additional metadata or notes (optional)
}

// TextFilterRequest represents the request to get a text secret by its name.
type TextFilterRequest struct {
	SecretName string `json:"secret_name,omitempty"` // Unique name of the secret to retrieve
}

// TextDB represents a stored text secret.
type TextDB struct {
	SecretName  string            `json:"secret_name" db:"secret_name"`   // Unique name of the secret
	SecretOwner string            `json:"secret_owner" db:"secret_owner"` // Owner of the secret
	Content     string            `json:"content" db:"content"`           // Text content of the secret
	Meta        *fields.StringMap `json:"meta,omitempty" db:"meta"`       // Optional metadata or notes
	UpdatedAt   time.Time         `json:"updated_at" db:"updated_at"`     // Timestamp of last update
}
