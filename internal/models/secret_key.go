package models

import "time"

// SecretKeyResponse описывает JSON-ответ с данными секретного ключа.
// swagger:model SecretKeyResponse
type SecretKeyRequest struct {
	// Идентификатор секрета
	// example: "secret-12345"
	// default: "secret-12345"
	SecretID string `json:"secret_id"`
	// Идентификатор устройства
	// example: "device-67890"
	// default: "device-67890"
	DeviceID string `json:"device_id"`
	// AES ключ, зашифрованный публичным ключом устройства
	// example: "U2FsdGVkX1+abcd1234efgh5678ijkl90=="
	// default: "U2FsdGVkX1+abcd1234efgh5678ijkl90=="
	EncryptedAESKey []byte `json:"encrypted_aes_key"`
}

// SecretKeyResponse описывает JSON-ответ с данными секретного ключа.
// swagger:model SecretKeyResponse
type SecretKeyResponse struct {
	// Уникальный идентификатор записи секретного ключа
	// example: "a1b2c3d4-e5f6-7890-abcd-1234567890ef"
	// default: "a1b2c3d4-e5f6-7890-abcd-1234567890ef"
	SecretKeyID string `json:"secret_key_id"`
	// Идентификатор секрета
	// example: "secret-12345"
	// default: "secret-12345"
	SecretID string `json:"secret_id"`
	// Идентификатор устройства
	// example: "device-67890"
	// default: "device-67890"
	DeviceID string `json:"device_id"`
	// AES ключ, зашифрованный публичным ключом устройства
	// example: "U2FsdGVkX1+abcd1234efgh5678ijkl90=="
	// default: "U2FsdGVkX1+abcd1234efgh5678ijkl90=="
	EncryptedAESKey string `json:"encrypted_aes_key"`
	// Дата создания записи
	// example: 2025-08-17T12:34:56Z
	// default: 2025-08-17T12:34:56Z
	CreatedAt time.Time `json:"created_at"`
	// Дата последнего обновления записи
	// example: 2025-08-17T12:45:00Z
	// default: 2025-08-17T12:45:00Z
	UpdatedAt time.Time `json:"updated_at"`
}

// SecretKeyDB представляет запись зашифрованного AES ключа для устройства
type SecretKeyDB struct {
	SecretKeyID     string    `json:"secret_key_id" db:"secret_key_id"`         // уникальный идентификатор записи
	SecretID        string    `json:"secret_id" db:"secret_id"`                 // идентификатор секрета
	DeviceID        string    `json:"device_id" db:"device_id"`                 // идентификатор устройства
	EncryptedAESKey []byte    `json:"encrypted_aes_key" db:"encrypted_aes_key"` // AES ключ, зашифрованный публичным ключом устройства
	CreatedAt       time.Time `json:"created_at" db:"created_at"`               // дата создания записи
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`               // дата последнего обновления записи
}
