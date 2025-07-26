package models

// UserRegisterRequest represents a request message for user registration.
type UserRegisterRequest struct {
	Username         string `json:"username"`
	Password         string `json:"password"`
	ClientPubKeyFile string `json:"client_pub_key_file"`
}

// UserRegisterResponse represents a response message after successful registration.
type UserRegisterResponse struct {
	Token string `json:"token"`
}

// UserLoginRequest represents a request message for user registration.
type UserLoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// UserLoginResponse represents a response message after successful registration.
type UserLoginResponse struct {
	Token string `json:"token"`
}

// User represents domain entity
type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
