package models

import "time"

// SecretBankCardSaveRequest содержит данные для сохранения банковской карты.
type SecretBankCardSaveRequest struct {
	SecretName string  `json:"secret_name" db:"secret_name"`
	Number     string  `json:"number" db:"number"`
	Owner      string  `json:"owner" db:"owner"`
	Exp        string  `json:"exp" db:"exp"`
	CVV        string  `json:"cvv" db:"cvv"`
	Meta       *string `json:"meta,omitempty" db:"meta"`
}

// SecretBankCardGetResponse представляет данные, возвращаемые при получении банковской карты.
type SecretBankCardGetResponse struct {
	SecretName  string     `json:"secret_name" db:"secret_name"`   // Уникальное имя секрета.
	SecretOwner string     `json:"secret_owner" db:"secret_owner"` // Идентификатор владельца секрета.
	Number      string     `json:"number" db:"number"`             // Номер банковской карты (PAN).
	Owner       string     `json:"owner" db:"owner"`               // Имя владельца карты (опционально).
	Exp         string     `json:"exp" db:"exp"`                   // Срок действия карты в формате MM/YY.
	CVV         string     `json:"cvv" db:"cvv"`                   // Код безопасности (3 или 4 цифры).
	Meta        *string    `json:"meta,omitempty" db:"meta"`       // Дополнительные метаданные (JSON).
	UpdatedAt   *time.Time `json:"updated_at" db:"updated_at"`     // Время последнего обновления.
}

// SecretBankCardListResponse представляет список банковских карт.
type SecretBankCardListResponse struct {
	Items []SecretBankCardGetResponse `json:"items"` // Список карт.
}

// SecretBankCardGetRequest содержит имя секрета, который необходимо получить.
type SecretBankCardGetRequest struct {
	SecretName string `json:"secret_name" db:"secret_name"` // Уникальное имя секрета.
}

// SecretBankCardDB представляет данные в бд
type SecretBankCardDB struct {
	SecretName  string     `json:"secret_name" db:"secret_name"`   // Уникальное имя секрета.
	SecretOwner string     `json:"secret_owner" db:"secret_owner"` // Идентификатор владельца секрета.
	Number      string     `json:"number" db:"number"`             // Номер банковской карты (PAN).
	Owner       string     `json:"owner" db:"owner"`               // Имя владельца карты (опционально).
	Exp         string     `json:"exp" db:"exp"`                   // Срок действия карты в формате MM/YY.
	CVV         string     `json:"cvv" db:"cvv"`                   // Код безопасности (3 или 4 цифры).
	Meta        *string    `json:"meta,omitempty" db:"meta"`       // Дополнительные метаданные (JSON).
	UpdatedAt   *time.Time `json:"updated_at" db:"updated_at"`     // Время последнего обновления.
}
