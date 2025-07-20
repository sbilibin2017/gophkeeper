package models

// TextAddRequest represents the request payload for adding a new text secret.
type TextAddRequest struct {
	SecretName string  `json:"secret_name,omitempty" validate:"required"` // Unique name of the secret
	Content    string  `json:"content,omitempty" validate:"required"`     // Text content of the secret
	Meta       *string `json:"meta,omitempty"`                            // Additional metadata or notes (optional)
}

// TextGetRequest represents the request to get a text secret by its name.
type TextGetRequest struct {
	SecretName string `json:"secret_name,omitempty"` // Unique name of the secret to retrieve
}

// TextDeleteRequest represents the request to delete a text secret by its name.
type TextDeleteRequest struct {
	SecretName string `json:"secret_name,omitempty"` // Unique name of the secret to delete
}

// TextDB represents the database model for a stored text secret.
type TextDB struct {
	SecretName  string  `json:"secret_name,omitempty" db:"secret_name"`   // Unique name of the secret
	SecretOwner string  `json:"secret_owner,omitempty" db:"secret_owner"` // Owner of the secret (user)
	Content     string  `json:"content,omitempty" db:"content"`           // Text content of the secret
	Meta        *string `json:"meta,omitempty" db:"meta"`                 // Additional metadata or notes
	UpdatedAt   string  `json:"updated_at,omitempty" db:"updated_at"`     // Timestamp of last update
}
