package models

import "time"

// SecretUsernamePasswordSaveRequest содержит данные для сохранения секрета с логином и паролем.
type SecretUsernamePasswordSaveRequest struct {
	SecretName string  `json:"secret_name" db:"secret_name"` // Уникальное имя секрета в рамках владельца.
	Username   string  `json:"username" db:"username"`       // Имя пользователя.
	Password   string  `json:"password" db:"password"`       // Пароль.
	Meta       *string `json:"meta,omitempty" db:"meta"`     // Дополнительные метаданные в формате JSON (может быть nil).
}

// SecretUsernamePasswordGetRequest содержит имя секрета, который необходимо получить.
type SecretUsernamePasswordGetRequest struct {
	SecretName string `json:"secret_name" db:"secret_name"` // Уникальное имя секрета.
}

// SecretUsernamePasswordGetResponse представляет данные, возвращаемые при получении секрета с логином и паролем.
type SecretUsernamePasswordGetResponse struct {
	SecretName  string     `json:"secret_name" db:"secret_name"`   // Уникальное имя секрета.
	SecretOwner string     `json:"secret_owner" db:"secret_owner"` // Идентификатор владельца секрета.
	Username    string     `json:"username" db:"username"`         // Имя пользователя.
	Password    string     `json:"password" db:"password"`         // Пароль.
	Meta        *string    `json:"meta,omitempty" db:"meta"`       // Дополнительные метаданные.
	UpdatedAt   *time.Time `json:"updated_at" db:"updated_at"`     // Время последнего обновления.
}

// SecretUsernamePasswordListResponse представляет список секретов с логином и паролем.
type SecretUsernamePasswordListResponse struct {
	Items []SecretUsernamePasswordGetResponse `json:"items"` // Список секретов.
}

// SecretUsernamePasswordDB представляет данные секрета с логином и паролем в базе данных.
type SecretUsernamePasswordDB struct {
	SecretName  string     `json:"secret_name" db:"secret_name"`   // Уникальное имя секрета в рамках владельца.
	SecretOwner string     `json:"secret_owner" db:"secret_owner"` // Идентификатор владельца секрета.
	Username    string     `json:"username" db:"username"`         // Имя пользователя.
	Password    string     `json:"password" db:"password"`         // Пароль.
	Meta        *string    `json:"meta,omitempty" db:"meta"`       // Дополнительные метаданные.
	UpdatedAt   *time.Time `json:"updated_at" db:"updated_at"`     // Время последнего обновления.
}
