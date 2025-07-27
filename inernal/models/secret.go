package models

import (
	"time"
)

const (
	SecretTypeBankCard = "bankcard" // Secret type: BankCard
	SecretTypeUser     = "user"     // Secret type: User
	SecretTypeText     = "text"     // Secret type: Text
	SecretTypeBinary   = "binary"   // Secret type: Binary
)

// Secret — структура для хранения секрета в БД
type Secret struct {
	SecretName  string    `db:"secret_name" json:"secret_name"`
	SecretType  string    `db:"secret_type" json:"secret_type"`
	SecretOwner string    `db:"secret_owner" json:"secret_owner"`
	Ciphertext  []byte    `db:"ciphertext" json:"ciphertext"`
	AESKeyEnc   []byte    `db:"aes_key_enc" json:"aes_key_enc"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`
}

type SecretEncrypted struct {
	Ciphertext []byte
	AESKeyEnc  []byte
}

type Bankcard struct {
	SecretName string  `json:"secret_name"`
	Number     string  `json:"number"`
	Owner      string  `json:"owner"`
	Exp        string  `json:"exp"`
	CVV        string  `json:"cvv"`
	Meta       *string `json:"meta,omitempty"`
}

type Text struct {
	SecretName string  `json:"secret_name"`
	Data       string  `json:"data"`
	Meta       *string `json:"meta,omitempty"`
}

type Binary struct {
	SecretName string  `json:"secret_name"`
	Data       []byte  `json:"data"`
	Meta       *string `json:"meta,omitempty"`
}

type User struct {
	SecretName string  `json:"secret_name"`
	Username   string  `json:"username"`
	Password   string  `json:"password"`
	Meta       *string `json:"meta,omitempty"`
}
