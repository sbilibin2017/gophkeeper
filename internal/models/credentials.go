package models

// Credentials represents user credentials
// used for authentication (e.g., login or registration).
type Credentials struct {
	Username string `json:"username"` // username
	Password string `json:"password"` // user password
}
