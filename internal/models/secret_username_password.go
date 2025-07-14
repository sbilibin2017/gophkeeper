package models

import "time"

// SecretUsernamePasswordClient представляет секрет с логином и паролем.
type SecretUsernamePasswordClient struct {
	SecretName string    `json:"secret_name" db:"secret_name"` // Уникальное имя секрета
	Username   string    `json:"username" db:"username"`       // Имя пользователя
	Password   string    `json:"password" db:"password"`       // Пароль
	Meta       *string   `json:"meta,omitempty" db:"meta"`     // Дополнительные данные в формате JSON, может быть nil
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`   // Время последнего обновления
}
