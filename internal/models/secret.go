package models

import "time"

// SecretRequest описывает тело запроса для сохранения секрета.
// swagger:model SecretRequest
type SecretRequest struct {
	// Идентификатор пользователя
	// example: "user789"
	// default: "user789"
	UserID string `json:"user_id"`
	// Название секрета
	// example: "my-password"
	// default: "my-password"
	SecretName string `json:"secret_name"`
	// Тип секрета
	// example: "password"
	// default: "password"
	SecretType string `json:"secret_type"`
	// Зашифрованное содержимое секрета
	// example: "SGVsbG8gV29ybGQh"
	// default: "SGVsbG8gV29ybGQh"
	EncryptedPayload string `json:"encrypted_payload"`
	// Nonce для шифрования
	// example: "MTIzNDU2Nzg5MA=="
	// default: "MTIzNDU2Nzg5MA=="
	Nonce string `json:"nonce"`
	// Метаданные секрета в формате JSON
	// example: {"url":"https://example.com"}
	// default: "{\"url\":\"https://example.com\"}"
	Meta string `json:"meta"`
}

// SecretResponse описывает JSON-ответ с данными секрета.
// swagger:model SecretResponse
type SecretResponse struct {
	// Уникальный идентификатор секрета
	// example: "abc123"
	// default: "abc123"
	SecretID string `json:"secret_id"`
	// Идентификатор пользователя
	// example: "user789"
	// default: "user789"
	UserID string `json:"user_id"`
	// Название секрета
	// example: "MyBankPassword"
	// default: "MyBankPassword"
	SecretName string `json:"secret_name"`
	// Тип секрета
	// example: "password"
	// default: "password"
	SecretType string `json:"secret_type"`
	// Зашифрованное содержимое секрета
	// example: "U2FsdGVkX1+abc123xyz=="
	// default: "U2FsdGVkX1+abc123xyz=="
	EncryptedPayload string `json:"encrypted_payload"`
	// Nonce для шифрования
	// example: "bXlOb25jZQ=="
	// default: "bXlOb25jZQ=="
	Nonce string `json:"nonce"`
	// Метаданные секрета в формате JSON
	// example: "{\"url\":\"https://example.com\",\"note\":\"для личного пользования\"}"
	// default: "{\"url\":\"https://example.com\",\"note\":\"для личного пользования\"}"
	Meta string `json:"meta"`
	// Дата создания секрета
	// example: "2025-08-17T12:00:00Z"
	// default: "2025-08-17T12:00:00Z"
	CreatedAt time.Time `json:"created_at"`
	// Дата последнего обновления секрета
	// example: "2025-08-17T12:30:00Z"
	// default: "2025-08-17T12:30:00Z"
	UpdatedAt time.Time `json:"updated_at"`
}

// SecretDB представляет запись секрета пользователя
type SecretDB struct {
	SecretID         string    `json:"secret_id" db:"secret_id"`                 // уникальный идентификатор секрета
	UserID           string    `json:"user_id" db:"user_id"`                     // идентификатор пользователя-владельца секрета
	SecretName       string    `json:"secret_name" db:"secret_name"`             // человекочитаемое имя секрета
	SecretType       string    `json:"secret_type" db:"secret_type"`             // тип секрета: password, card, note и т.д.
	EncryptedPayload string    `json:"encrypted_payload" db:"encrypted_payload"` // зашифрованные данные AES
	Nonce            string    `json:"nonce" db:"nonce"`                         // nonce для AES-GCM
	Meta             string    `json:"meta" db:"meta"`                           // JSON метаданные
	CreatedAt        time.Time `json:"created_at" db:"created_at"`               // дата создания секрета
	UpdatedAt        time.Time `json:"updated_at" db:"updated_at"`               // дата последнего обновления секрета
}

// BankcardPayload represents a bank card secret payload.
type BankcardPayload struct {
	Number string  `json:"number"`
	Owner  string  `json:"owner"`
	Exp    string  `json:"exp"`
	CVV    string  `json:"cvv"`
	Meta   *string `json:"meta,omitempty"`
}

// TextPayload represents a text secret payload.
type TextPayload struct {
	Data string  `json:"data"`
	Meta *string `json:"meta,omitempty"`
}

// BinaryPayload represents a binary secret payload.
type BinaryPayload struct {
	Data []byte  `json:"data"`
	Meta *string `json:"meta,omitempty"`
}

// UserPayload represents a user secret payload.
type UserPayload struct {
	Username string  `json:"username"`
	Password string  `json:"password"`
	Meta     *string `json:"meta,omitempty"`
}
