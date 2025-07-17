package models

// AuthRequest represents an authentication request with username and password.
type AuthRequest struct {
	Username string `json:"username"` // Username for authentication
	Password string `json:"password"` // Password for authentication
}

// AuthResponse represents an authentication response containing a token.
type AuthResponse struct {
	Token string `json:"token"` // JWT or authentication token
}
