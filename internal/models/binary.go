package models

// AddSecretBinaryRequest represents a request to add a binary secret.
type AddSecretBinaryRequest struct {
	SecretName string  `json:"secret_name"`    // Secret name
	Data       []byte  `json:"data"`           // Binary data
	Meta       *string `json:"meta,omitempty"` // Optional metadata
}

// GetSecretBinaryRequest represents a request to get a binary secret.
type GetSecretBinaryRequest struct {
	SecretName string `json:"secret_name"` // Secret name
}

// GetSecretBinaryResponse represents a response with a binary secret.
type GetSecretBinaryResponse struct {
	SecretName  string  `json:"secret_name"`    // Secret name
	SecretOwner string  `json:"secret_owner"`   // Secret owner
	Data        []byte  `json:"data"`           // Binary data
	Meta        *string `json:"meta,omitempty"` // Optional metadata
	UpdatedAt   string  `json:"updated_at"`     // Last update timestamp
}

// ListSecretBinaryResponse contains a list of binary secrets.
type ListSecretBinaryResponse struct {
	Items []GetSecretBinaryResponse `json:"items"` // List of binary secrets
}
