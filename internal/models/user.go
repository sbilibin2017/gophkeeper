package models

import (
	"time"
)

// AuthRequest represents the request payload for user authentication or registration.
type AuthRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// AuthResponse represents the response payload from authentication or registration.
type AuthResponse struct {
	Token string `json:"token"`
}

// User represents a domain.
type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// UserDB represents a user record in the database.
type UserDB struct {
	Username     string    `json:"username" db:"username"`
	PasswordHash string    `json:"password_hash" db:"password_hash"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}
