package models

// AuthRequest represents the request payload for user authentication.
// It contains the login credentials: username and password.
type AuthRequest struct {
	Login    string `json:"login"`    // Username or login identifier
	Password string `json:"password"` // User password
}

// AuthResponse represents the response payload for authentication requests.
// It contains the authentication token.
type AuthResponse struct {
	Token string `json:"token"` // Authentication token
}
