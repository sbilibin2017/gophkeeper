package models

import "time"

// SecretTextSaveRequest содержит данные для сохранения текстового секрета.
type SecretTextSaveRequest struct {
	SecretName string  `json:"secret_name" db:"secret_name"` // Уникальное имя секрета в рамках владельца.
	Content    string  `json:"content" db:"content"`         // Текстовое содержимое секрета.
	Meta       *string `json:"meta,omitempty" db:"meta"`     // Дополнительные метаданные в формате JSON (может быть nil).
}

// SecretTextGetRequest содержит имя текстового секрета, который необходимо получить.
type SecretTextGetRequest struct {
	SecretName string `json:"secret_name" db:"secret_name"` // Уникальное имя секрета.
}

// SecretTextGetResponse представляет данные, возвращаемые при получении текстового секрета.
type SecretTextGetResponse struct {
	SecretName  string     `json:"secret_name" db:"secret_name"`   // Уникальное имя секрета.
	SecretOwner string     `json:"secret_owner" db:"secret_owner"` // Идентификатор владельца секрета.
	Content     string     `json:"content" db:"content"`           // Текстовое содержимое секрета.
	Meta        *string    `json:"meta,omitempty" db:"meta"`       // Дополнительные метаданные.
	UpdatedAt   *time.Time `json:"updated_at" db:"updated_at"`     // Время последнего обновления.
}

// SecretTextListResponse представляет список текстовых секретов.
type SecretTextListResponse struct {
	Items []SecretTextGetResponse `json:"items"` // Список текстовых секретов.
}

// SecretTextDB представляет данные текстового секрета в базе данных.
type SecretTextDB struct {
	SecretName  string     `json:"secret_name" db:"secret_name"`   // Уникальное имя секрета в рамках владельца.
	SecretOwner string     `json:"secret_owner" db:"secret_owner"` // Идентификатор владельца секрета.
	Content     string     `json:"content" db:"content"`           // Текстовое содержимое секрета.
	Meta        *string    `json:"meta,omitempty" db:"meta"`       // Дополнительные метаданные.
	UpdatedAt   *time.Time `json:"updated_at" db:"updated_at"`     // Время последнего обновления.
}
