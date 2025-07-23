package models

import "time"

// Binary represents a binary secret, including file path and data content.
// It holds optional metadata and tracks when it was last updated.
type Binary struct {
	SecretName  string    `json:"secret_name" db:"secret_name"`   // Unique identifier for the secret
	SecretOwner string    `json:"secret_owner" db:"secret_owner"` // Owner of the secret (e.g., user ID or username)
	FilePath    string    `json:"file_path" db:"file_path"`       // Path to the binary file
	Data        []byte    `json:"data" db:"data"`                 // Binary data content
	Meta        *string   `json:"meta,omitempty" db:"meta"`       // Optional metadata
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`     // Timestamp of the last update
}

// GetSecretName returns the unique secret name of the binary data.
func (b *Binary) GetSecretName() string {
	return b.SecretName
}

// GetUpdatedAt returns the timestamp when the binary secret was last updated.
func (b *Binary) GetUpdatedAt() time.Time {
	return b.UpdatedAt
}

// BinaryData contains the actual binary data, file path, optional metadata, and last update timestamp.
// This struct can be used separately from Binary when you don't need to include secret ownership information.
type BinaryData struct {
	FilePath string  `json:"file_path"`      // Path to the binary file
	Data     []byte  `json:"data"`           // Binary data content
	Meta     *string `json:"meta,omitempty"` // Optional metadata
}
