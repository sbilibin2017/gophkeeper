package models

import "time"

// BinaryAddRequest represents a request to add a binary secret.
type BinaryAddRequest struct {
	SecretName string  `json:"secret_name"`    // Secret name
	Data       []byte  `json:"data"`           // Binary data to store
	Meta       *string `json:"meta,omitempty"` // Optional metadata
}

// BinaryGetRequest represents a request to retrieve a binary secret by name.
type BinaryGetRequest struct {
	SecretName string `json:"secret_name"` // Secret name
}

// BinaryResponse contains the retrieved binary secret and metadata.
type BinaryResponse struct {
	SecretName  string    `json:"secret_name"`    // Secret name
	SecretOwner string    `json:"secret_owner"`   // Username of the secret's owner
	Data        []byte    `json:"data"`           // Stored binary data
	Meta        *string   `json:"meta,omitempty"` // Optional metadata
	UpdatedAt   time.Time `json:"updated_at"`     // Last modification timestamp
}

// BinaryListResponse contains a list of all binary secrets owned or visible to the user.
type BinaryListResponse struct {
	Items []BinaryResponse `json:"items"` // List of binary secrets
}
