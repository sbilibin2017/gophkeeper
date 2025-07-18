package models

import "time"

// UsernamePasswordAddRequest represents a request to add a login/password secret.
type UsernamePasswordAddRequest struct {
	SecretName string  `json:"secret_name" db:"secret_name"` // Secret name
	Username   string  `json:"username" db:"username"`       // Login username
	Password   string  `json:"password" db:"password"`       // Login password
	Meta       *string `json:"meta,omitempty" db:"meta"`     // Optional metadata
}

// UsernamePasswordGetRequest represents a request to retrieve a username-password secret by name.
type UsernamePasswordGetRequest struct {
	SecretName string `json:"secret_name"` // Secret name
}

// UsernamePasswordResponse contains the retrieved username-password secret and metadata.
type UsernamePasswordResponse struct {
	SecretName  string    `json:"secret_name"`    // Secret name
	SecretOwner string    `json:"secret_owner"`   // Username of the secret's owner
	Username    string    `json:"username"`       // Login username
	Password    string    `json:"password"`       // Login password
	Meta        *string   `json:"meta,omitempty"` // Optional metadata
	UpdatedAt   time.Time `json:"updated_at"`     // Last modification timestamp
}
