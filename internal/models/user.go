package models

import "time"

// User represents a user secret containing login credentials and optional metadata.
// It includes the secret name, owner information, login, password, optional metadata,
// and the timestamp when it was last updated.
type User struct {
	SecretName  string    `json:"secret_name" db:"secret_name"`   // Name identifier for the secret
	SecretOwner string    `json:"secret_owner" db:"secret_owner"` // Owner of the secret (e.g., user ID or username)
	Login       string    `json:"login" db:"login"`               // User login name
	Password    string    `json:"password" db:"password"`         // User password
	Meta        *string   `json:"meta,omitempty" db:"meta"`       // Optional metadata about the user secret
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`     // Timestamp of the last update
}

// GetSecretName returns the secret name of the User.
func (u *User) GetSecretName() string {
	return u.SecretName
}

// GetUpdatedAt returns the last updated time of the User.
func (u *User) GetUpdatedAt() time.Time {
	return u.UpdatedAt
}

// UserData contains login credentials and optional metadata for a user secret.
// This struct is used when ownership information is not required.
type UserData struct {
	Login    string  `json:"login"`          // User login name
	Password string  `json:"password"`       // User password
	Meta     *string `json:"meta,omitempty"` // Optional metadata about the user secret
}
