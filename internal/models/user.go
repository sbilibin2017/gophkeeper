package models

import "time"

// UserDB представляет таблицу users
type UserDB struct {
	UserID       string    `json:"user_id" db:"user_id"`             // уникальный идентификатор пользователя (UUID)
	Username     string    `json:"username" db:"username"`           // логин пользователя
	PasswordHash string    `json:"password_hash" db:"password_hash"` // хэш пароля
	CreatedAt    time.Time `json:"created_at" db:"created_at"`       // дата создания записи
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`       // дата последнего обновления
}
