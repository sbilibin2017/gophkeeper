package models

import "time"

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
