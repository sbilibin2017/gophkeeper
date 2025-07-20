package models

// BinaryAddRequest represents the request payload for adding a new binary secret.
type BinaryAddRequest struct {
	SecretName string  `json:"secret_name,omitempty" validate:"required"` // Unique name of the secret
	Data       []byte  `json:"data,omitempty" validate:"required"`        // Binary data of the secret
	Meta       *string `json:"meta,omitempty"`                            // Additional metadata or notes (optional)
}

// BinaryGetRequest represents the request to get a binary secret by its name.
type BinaryGetRequest struct {
	SecretName string `json:"secret_name,omitempty"` // Unique name of the secret to retrieve
}

// BinaryDeleteRequest represents the request to delete a binary secret by its name.
type BinaryDeleteRequest struct {
	SecretName string `json:"secret_name,omitempty"` // Unique name of the secret to delete
}

// BinaryDB represents the database model for a stored binary secret.
type BinaryDB struct {
	SecretName  string  `json:"secret_name,omitempty" db:"secret_name"`   // Unique name of the secret
	SecretOwner string  `json:"secret_owner,omitempty" db:"secret_owner"` // Owner of the secret (user)
	Data        []byte  `json:"data,omitempty" db:"data"`                 // Binary data of the secret
	Meta        *string `json:"meta,omitempty" db:"meta"`                 // Additional metadata or notes
	UpdatedAt   string  `json:"updated_at,omitempty" db:"updated_at"`     // Timestamp of last update
}
