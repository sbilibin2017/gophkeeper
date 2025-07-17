package models

// RegisterRequest mirrors the gRPC RegisterRequest message.
type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// RegisterResponse mirrors the gRPC RegisterResponse message.
type RegisterResponse struct {
	Token string `json:"token"`
}

// LoginRequest mirrors the gRPC LoginRequest message.
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginResponse mirrors the gRPC LoginResponse message.
type LoginResponse struct {
	Token string `json:"token"`
}

// LogoutRequest mirrors the gRPC LogoutRequest message.
type LogoutRequest struct {
	Token string `json:"token"`
}
