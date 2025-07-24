package models

// AuthRequest represents a user authentication request.
// It contains the user's login, password, and public key.
type AuthRequest struct {
	Login     string `json:"login"`      // The user's login or identifier
	Password  string `json:"password"`   // The user's password
	PublicKey string `json:"public_key"` // The user's public key (in PEM or base64 format)
}

// AuthResponse represents the response to an authentication request.
// It contains an authentication token.
type AuthResponse struct {
	Token string `json:"token"` // The authentication token
}
