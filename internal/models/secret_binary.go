package models

import "time"

// SecretBinarySaveRequest содержит данные для сохранения бинарного секрета.
type SecretBinarySaveRequest struct {
	SecretName string  `json:"secret_name" db:"secret_name"` // Уникальное имя секрета в рамках владельца.
	Data       []byte  `json:"data" db:"data"`               // Бинарные данные (например, файл, сертификат и т.п.).
	Meta       *string `json:"meta,omitempty" db:"meta"`     // Дополнительные метаданные в формате JSON (может быть nil).
}

// SecretBinaryGetRequest содержит имя бинарного секрета, который необходимо получить.
type SecretBinaryGetRequest struct {
	SecretName string `json:"secret_name" db:"secret_name"` // Уникальное имя секрета.
}

// SecretBinaryGetResponse представляет данные, возвращаемые при получении бинарного секрета.
type SecretBinaryGetResponse struct {
	SecretName  string     `json:"secret_name" db:"secret_name"`   // Уникальное имя секрета.
	SecretOwner string     `json:"secret_owner" db:"secret_owner"` // Идентификатор владельца секрета.
	Data        []byte     `json:"data" db:"data"`                 // Содержимое бинарного секрета.
	Meta        *string    `json:"meta,omitempty" db:"meta"`       // Дополнительные метаданные.
	UpdatedAt   *time.Time `json:"updated_at" db:"updated_at"`     // Время последнего обновления.
}

// SecretBinaryListResponse представляет список бинарных секретов.
type SecretBinaryListResponse struct {
	Items []SecretBinaryGetResponse `json:"items"` // Список бинарных секретов.
}

// SecretBinaryDB представляет данные бинарного секрета в базе данных.
type SecretBinaryDB struct {
	SecretName  string     `json:"secret_name" db:"secret_name"`   // Уникальное имя секрета в рамках владельца.
	SecretOwner string     `json:"secret_owner" db:"secret_owner"` // Идентификатор владельца секрета.
	Data        []byte     `json:"data" db:"data"`                 // Бинарные данные (например, файл, сертификат и т.п.).
	Meta        *string    `json:"meta,omitempty" db:"meta"`       // Дополнительные метаданные в формате JSON.
	UpdatedAt   *time.Time `json:"updated_at" db:"updated_at"`     // Время последнего обновления.
}
