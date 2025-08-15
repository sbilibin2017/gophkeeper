package models

import "time"

// DeviceDB представляет таблицу devices
type DeviceDB struct {
	DeviceID     string    `json:"device_id" db:"device_id"`         // уникальный идентификатор устройства (UUID)
	UserID       string    `json:"user_id" db:"user_id"`             // идентификатор пользователя-владельца
	DeviceName   string    `json:"device_name" db:"device_name"`     // название устройства, например "iPhone" или "PC"
	PublicKey    string    `json:"public_key" db:"public_key"`       // публичный ключ устройства
	EncryptedDEK string    `json:"encrypted_dek" db:"encrypted_dek"` // DEK зашифрованный публичным ключом устройства
	CreatedAt    time.Time `json:"created_at" db:"created_at"`       // дата создания записи
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`       // дата последнего обновления
}
