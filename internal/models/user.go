package models

// UserAddRequest represents the request payload for adding a new user secret.
type UserAddRequest struct {
	SecretName string  `json:"secret_name,omitempty" validate:"required"`                 // Unique name of the secret
	Username   string  `json:"username,omitempty" validate:"required,min=3,max=30,alpha"` // Username must be letters only, 3-30 chars
	Password   string  `json:"password,omitempty" validate:"required,min=8,max=128"`      // Password must be 8-128 chars
	Meta       *string `json:"meta,omitempty"`                                            // Additional metadata or notes (optional)
}

// UserGetRequest represents the request to get a user secret by its name.
type UserGetRequest struct {
	SecretName string `json:"secret_name,omitempty"` // Unique name of the secret to retrieve
}

// UserDeleteRequest represents the request to delete a user secret by its name.
type UserDeleteRequest struct {
	SecretName string `json:"secret_name,omitempty"` // Unique name of the secret to delete
}

// UserDB represents the database model for a stored user secret.
type UserDB struct {
	SecretName  string  `json:"secret_name,omitempty" db:"secret_name"`   // Unique name of the secret
	SecretOwner string  `json:"secret_owner,omitempty" db:"secret_owner"` // Owner of the secret (user)
	Username    string  `json:"username,omitempty" db:"username"`         // Username associated with the secret
	Password    string  `json:"password,omitempty" db:"password"`         // Password associated with the secret
	Meta        *string `json:"meta,omitempty" db:"meta"`                 // Additional metadata or notes
	UpdatedAt   string  `json:"updated_at,omitempty" db:"updated_at"`     // Timestamp of last update
}
