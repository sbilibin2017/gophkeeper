package models

import "time"

// SecretKeyDB представляет запись зашифрованного AES ключа для устройства
type SecretKeyDB struct {
	SecretKeyID     string    `json:"secret_key_id" db:"secret_key_id"`         // уникальный идентификатор записи
	SecretID        string    `json:"secret_id" db:"secret_id"`                 // идентификатор секрета
	DeviceID        string    `json:"device_id" db:"device_id"`                 // идентификатор устройства
	EncryptedAESKey string    `json:"encrypted_aes_key" db:"encrypted_aes_key"` // AES ключ, зашифрованный публичным ключом устройства
	CreatedAt       time.Time `json:"created_at" db:"created_at"`               // дата создания записи
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`               // дата последнего обновления записи
}
