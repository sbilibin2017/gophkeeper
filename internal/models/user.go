package models

import "time"

// User represents a user account in the system.
type User struct {
	Username     string    `json:"username" db:"username"`           // Username is the unique identifier for the user.
	PasswordHash string    `json:"password_hash" db:"password_hash"` // PasswordHash is the hashed password.
	CreatedAt    time.Time `json:"created_at" db:"created_at"`       // CreatedAt is when the user was created.
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`       // UpdatedAt is the last update time.
}
