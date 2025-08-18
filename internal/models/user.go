package models

import "time"

// RegisterRequest определяет входящий запрос на регистрацию пользователя.
// swagger:model RegisterRequest
type RegisterRequest struct {
	// Имя пользователя
	// required: true
	// example: johndoe
	Username string `json:"username"`
	// Пароль пользователя
	// required: true
	// example: Secret123!
	Password string `json:"password"`
}

// RegisterResponse определяет ответ на регистрацию пользователя.
// swagger:model RegisterResponse
type RegisterResponse struct {
	// Уникальный идентификатор пользователя
	// required: true
	// example: "c56a4180-65aa-42ec-a945-5fd21dec0538"
	UserID string `json:"user_id"`
	// Уникальный идентификатор устройства
	// required: true
	// example: "f47ac10b-58cc-4372-a567-0e02b2c3d479"
	DeviceID string `json:"device_id"`
	// Приватный ключ RSA (PEM кодирование)
	// required: true
	// example: |
	//   -----BEGIN RSA PRIVATE KEY-----
	//   MIIEpAIBAAKCAQEAu7pM4h2...
	//   -----END RSA PRIVATE KEY-----
	PrivateKey string `json:"private_key"`
}

// LoginRequest определяет входящий запрос на аутентификацию пользователя.
// swagger:model LoginRequest
type LoginRequest struct {
	// Имя пользователя
	// required: true
	// example: johndoe
	Username string `json:"username"`
	// Пароль пользователя
	// required: true
	// example: Secret123!
	Password string `json:"password"`
	// Уникальный идентификатор устройства
	// required: true
	// example: "f47ac10b-58cc-4372-a567-0e02b2c3d479"
	DeviceID string `json:"device_id"`
}

// User представляет запись пользователя
type UserDB struct {
	UserID       string    `json:"user_id" db:"user_id"`             // уникальный идентификатор пользователя
	Username     string    `json:"username" db:"username"`           // уникальное имя пользователя
	PasswordHash string    `json:"password_hash" db:"password_hash"` // хэш пароля
	CreatedAt    time.Time `json:"created_at" db:"created_at"`       // дата создания пользователя
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`       // дата последнего обновления данных пользователя
}
