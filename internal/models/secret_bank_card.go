package models

import "time"

// SecretBankCardClient представляет структуру банковской карты, хранимой в БД.
type SecretBankCardClient struct {
	SecretName string    `json:"secret_name" db:"secret_name"` // Название секрета (уникально в рамках владельца)
	Number     string    `json:"number" db:"number"`           // Номер банковской карты (PAN)
	Owner      string    `json:"owner" db:"owner"`             // Владелец секрета (логически необязательный)
	Exp        string    `json:"exp" db:"exp"`                 // Срок действия карты в формате MM/YY
	CVV        string    `json:"cvv" db:"cvv"`                 // Код безопасности (3 или 4 цифры)
	Meta       *string   `json:"meta,omitempty" db:"meta"`     // Дополнительные метаданные в формате JSON (может быть nil)
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`   // Время последнего обновления записи
}
