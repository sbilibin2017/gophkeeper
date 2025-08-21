package models

import "time"

// DeviceResponse описывает JSON-ответ с данными устройства.
// swagger:model DeviceResponse
type DeviceResponse struct {
	// Уникальный идентификатор устройства
	// example: "f47ac10b-58cc-4372-a567-0e02b2c3d479"
	// default: "f47ac10b-58cc-4372-a567-0e02b2c3d479"
	DeviceID string `json:"device_id"`
	// Идентификатор пользователя-владельца устройства
	// example: "c56a4180-65aa-42ec-a945-5fd21dec0538"
	// default: "c56a4180-65aa-42ec-a945-5fd21dec0538"
	UserID string `json:"user_id"`
	// Публичный ключ устройства
	// example: |
	//   -----BEGIN PUBLIC KEY-----
	//   MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAu7pM4h2...
	//   -----END PUBLIC KEY-----
	// default: |
	//   -----BEGIN PUBLIC KEY-----
	//   MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAu7pM4h2...
	//   -----END PUBLIC KEY-----
	PublicKey string `json:"public_key"`
	// Дата создания устройства
	// example: 2025-08-17T12:34:56Z
	// default: 2025-08-17T12:34:56Z
	CreatedAt time.Time `json:"created_at"`
	// Дата последнего обновления данных устройства
	// example: 2025-08-17T12:45:00Z
	// default: 2025-08-17T12:45:00Z
	UpdatedAt time.Time `json:"updated_at"`
}

// DeviceDB представляет запись устройства пользователя
type DeviceDB struct {
	DeviceID  string    `json:"device_id" db:"device_id"`   // уникальный идентификатор устройства
	UserID    string    `json:"user_id" db:"user_id"`       // идентификатор пользователя-владельца устройства
	PublicKey string    `json:"public_key" db:"public_key"` // публичный ключ устройства
	CreatedAt time.Time `json:"created_at" db:"created_at"` // дата создания устройства
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"` // дата последнего обновления данных устройства
}
