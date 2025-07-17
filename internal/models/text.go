package models

import "time"

// TextAddRequest represents a request to add a plain text secret.
type TextAddRequest struct {
	SecretName string  `json:"secret_name"`    // Secret name
	Content    string  `json:"content"`        // Plain text content
	Meta       *string `json:"meta,omitempty"` // Optional metadata
}

// TextGetRequest represents a request to retrieve a text secret by name.
type TextGetRequest struct {
	SecretName string `json:"secret_name"` // Secret name
}

// TextResponse contains the retrieved text secret and metadata.
type TextResponse struct {
	SecretName  string    `json:"secret_name"`    // Secret name
	SecretOwner string    `json:"secret_owner"`   // Username of the secret's owner
	Content     string    `json:"content"`        // Plain text content
	Meta        *string   `json:"meta,omitempty"` // Optional metadata
	UpdatedAt   time.Time `json:"updated_at"`     // Last modification timestamp
}

// TextListResponse contains a list of all text secrets.
type TextListResponse struct {
	Items []TextResponse `json:"items"` // List of text secrets
}
