package models

// AuthRequest represents the request payload for user authentication.
// It contains the username and password credentials.
type AuthRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// AuthResponse represents the response returned after a successful authentication.
// It contains the authentication token.
type AuthResponse struct {
	Token string `json:"token"`
}
