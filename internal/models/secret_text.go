package models

import "time"

// SecretTextClient представляет структуру текстового секрета, хранимого на стороне клиента.
type SecretTextClient struct {
	SecretName string    `json:"secret_name" db:"secret_name"` // Название секрета (уникальное в рамках владельца)
	Content    string    `json:"content" db:"content"`         // Текстовое содержимое секрета
	Meta       *string   `json:"meta,omitempty" db:"meta"`     // Дополнительные метаданные в формате JSON (может быть nil)
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`   // Время последнего обновления секрета
}
