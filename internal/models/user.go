package models

import "time"

// UserDB представляет запись пользователя
type UserDB struct {
	UserID       string    `json:"user_id" db:"user_id"`             // уникальный идентификатор пользователя
	Username     string    `json:"username" db:"username"`           // уникальное имя пользователя
	PasswordHash string    `json:"password_hash" db:"password_hash"` // хэш пароля
	CreatedAt    time.Time `json:"created_at" db:"created_at"`       // дата создания пользователя
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`       // дата последнего обновления данных пользователя
}
