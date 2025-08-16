package models

import "time"

// DeviceDB представляет запись устройства пользователя
type DeviceDB struct {
	DeviceID  string    `json:"device_id" db:"device_id"`   // уникальный идентификатор устройства
	UserID    string    `json:"user_id" db:"user_id"`       // идентификатор пользователя-владельца устройства
	PublicKey string    `json:"public_key" db:"public_key"` // публичный ключ устройства
	CreatedAt time.Time `json:"created_at" db:"created_at"` // дата создания устройства
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"` // дата последнего обновления данных устройства
}
