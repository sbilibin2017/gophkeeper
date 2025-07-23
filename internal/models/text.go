package models

import "time"

// Text represents a textual secret with optional metadata and an update timestamp.
// It includes the secret name, owner information, the actual text data, optional metadata,
// and the time when it was last updated.
type Text struct {
	SecretName  string    `json:"secret_name" db:"secret_name"`   // Name identifier for the secret
	SecretOwner string    `json:"secret_owner" db:"secret_owner"` // Owner of the secret (e.g., user ID or username)
	Data        string    `json:"data" db:"data"`                 // The actual text data
	Meta        *string   `json:"meta,omitempty" db:"meta"`       // Optional metadata about the secret
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`     // Timestamp of the last update
}

// GetSecretName returns the secret name of the Text.
func (t *Text) GetSecretName() string {
	return t.SecretName
}

// GetUpdatedAt returns the last updated time of the Text.
func (t *Text) GetUpdatedAt() time.Time {
	return t.UpdatedAt
}

// TextData contains the textual secret data and optional metadata.
// This struct is used when ownership information is not required.
type TextData struct {
	Data string  `json:"data"`           // The actual text data
	Meta *string `json:"meta,omitempty"` // Optional metadata about the secret
}
