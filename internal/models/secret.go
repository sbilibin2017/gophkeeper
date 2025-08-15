package models

import "time"

// SecretDB представляет таблицу secrets
type SecretDB struct {
	SecretID   string    `json:"secret_id" db:"secret_id"`     // уникальный идентификатор секрета (UUID)
	UserID     string    `json:"user_id" db:"user_id"`         // идентификатор пользователя-владельца секрета
	SecretType string    `json:"secret_type" db:"secret_type"` // тип секрета: password, text, card, binary
	Title      string    `json:"title,omitempty" db:"title"`   // название или метка секрета
	Data       string    `json:"data" db:"data"`               // зашифрованные данные (Base64 или hex)
	Meta       *string   `json:"meta,omitempty" db:"meta"`     // метаинформация в формате JSON
	CreatedAt  time.Time `json:"created_at" db:"created_at"`   // дата создания записи
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`   // дата последнего обновления
}
