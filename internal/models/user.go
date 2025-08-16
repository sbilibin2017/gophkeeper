package models

import "time"

// RegisterRequest представляет тело запроса на регистрацию пользователя.
type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginRequest представляет тело запроса на аутентификацию пользователя.
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// RegisterResponse представляет тело ответа сервера на регистрацию пользователя.
type RegisterResponse struct {
	UserID     string `json:"user_id"`
	DeviceID   string `json:"device_id"`
	PrivateKey string `json:"private_key"`
	Token      string `json:"token"` // можно оставить пустым, если токен извлекается через TokenGetter
}

// LoginResponse представляет тело ответа сервера на вход пользователя.
type LoginResponse struct {
	Token string `json:"token"` // токен можно хранить здесь, либо извлекать через TokenGetter
}

// UserDB представляет запись пользователя
type UserDB struct {
	UserID       string    `json:"user_id" db:"user_id"`             // уникальный идентификатор пользователя
	Username     string    `json:"username" db:"username"`           // уникальное имя пользователя
	PasswordHash string    `json:"password_hash" db:"password_hash"` // хэш пароля
	CreatedAt    time.Time `json:"created_at" db:"created_at"`       // дата создания пользователя
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`       // дата последнего обновления данных пользователя
}
