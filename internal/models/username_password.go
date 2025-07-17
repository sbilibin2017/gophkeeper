package models

// UsernamePasswordAddRequest represents a request to add a username-password secret.
type UsernamePasswordAddRequest struct {
	SecretName string  `json:"secret_name"`    // Secret name
	Username   string  `json:"username"`       // Username
	Password   string  `json:"password"`       // Password
	Meta       *string `json:"meta,omitempty"` // Optional metadata
}

// UsernamePasswordGetRequest represents a request to get a username-password secret.
type UsernamePasswordGetRequest struct {
	SecretName string `json:"secret_name"` // Secret name
}

// UsernamePasswordResponse represents a response with a username-password secret.
type UsernamePasswordResponse struct {
	SecretName  string  `json:"secret_name"`    // Secret name
	SecretOwner string  `json:"secret_owner"`   // Secret owner
	Username    string  `json:"username"`       // Username
	Password    string  `json:"password"`       // Password
	Meta        *string `json:"meta,omitempty"` // Optional metadata
	UpdatedAt   string  `json:"updated_at"`     // Last update timestamp
}

// UsernamePasswordListResponse contains a list of username-password secrets.
type UsernamePasswordListResponse struct {
	Items []UsernamePasswordResponse `json:"items"` // List of username-password secrets
}
