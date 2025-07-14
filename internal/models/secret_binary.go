package models

import "time"

// SecretBinaryClient представляет структуру бинарного секрета, хранимого на стороне клиента.
type SecretBinaryClient struct {
	SecretName string    `json:"secret_name" db:"secret_name"` // Название секрета (уникальное в рамках владельца)
	Data       []byte    `json:"data" db:"data"`               // Бинарные данные (например, файл, сертификат и т.п.)
	Meta       *string   `json:"meta,omitempty" db:"meta"`     // Дополнительная мета-информация в формате JSON, может быть nil
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`   // Время последнего обновления секрета
}
