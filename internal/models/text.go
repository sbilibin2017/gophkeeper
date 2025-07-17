package models

// TextAddRequest represents a request to add a text secret.
type TextAddRequest struct {
	SecretName string  `json:"secret_name"`    // Secret name
	Content    string  `json:"content"`        // Text content
	Meta       *string `json:"meta,omitempty"` // Optional metadata
}

// TextGetRequest represents a request to get a text secret.
type TextGetRequest struct {
	SecretName string `json:"secret_name"` // Secret name
}

// TextResponse represents a response with text secret information.
type TextResponse struct {
	SecretName  string  `json:"secret_name"`    // Secret name
	SecretOwner string  `json:"secret_owner"`   // Secret owner
	Content     string  `json:"content"`        // Text content
	Meta        *string `json:"meta,omitempty"` // Optional metadata
	UpdatedAt   string  `json:"updated_at"`     // Last update timestamp
}

// TextListResponse contains a list of text secrets.
type TextListResponse struct {
	Items []TextResponse `json:"items"` // List of text secrets
}
