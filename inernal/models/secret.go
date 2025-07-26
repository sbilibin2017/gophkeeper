package models

import (
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	SecretTypeBankCard = "bankcard" // Secret type: BankCard
	SecretTypeUser     = "user"     // Secret type: User
	SecretTypeText     = "text"     // Secret type: Text
	SecretTypeBinary   = "binary"   // Secret type: Binary
)

const (
	SyncModeServer      = "server"      // Sync handled by the server
	SyncModeClient      = "client"      // Sync handled by the client
	SyncModeInteractive = "interactive" // Sync done interactively with user input
)

// SecretSaveRequest соответствует protobuf SecretSaveRequest
type SecretSaveRequest struct {
	SecretName string `json:"secret_name"`
	SecretType string `json:"secret_type"`
	Ciphertext []byte `json:"ciphertext"`
	AESKeyEnc  []byte `json:"aes_key_enc"`
	Token      string `json:"token"`
}

// SecretGetRequest соответствует protobuf SecretGetRequest
type SecretGetRequest struct {
	SecretName string `json:"secret_name"`
	SecretType string `json:"secret_type"`
	Token      string `json:"token"`
}

// SecretListRequest соответствует protobuf SecretListRequest
type SecretListRequest struct {
	Token string `json:"token"`
}

// SecretResponse соответствует protobuf SecretResponse
type SecretResponse struct {
	SecretName  string                 `json:"secret_name"`
	SecretType  string                 `json:"secret_type"`
	SecretOwner string                 `json:"secret_owner"`
	Ciphertext  []byte                 `json:"ciphertext"`
	AESKeyEnc   []byte                 `json:"aes_key_enc"`
	UpdatedAt   *timestamppb.Timestamp `json:"updated_at"`
}

// SecretDB — структура для хранения секрета в БД
type SecretDB struct {
	SecretName  string    `db:"secret_name" json:"secret_name"`
	SecretType  string    `db:"secret_type" json:"secret_type"`
	SecretOwner string    `db:"secret_owner" json:"secret_owner"`
	Ciphertext  []byte    `db:"ciphertext" json:"ciphertext"`
	AESKeyEnc   []byte    `db:"aes_key_enc" json:"aes_key_enc"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`
}

// SecretGetRequest соответствует protobuf SecretGetRequest
type SecretGetFilterDB struct {
	SecretName  string `json:"secret_name"`
	SecretType  string `json:"secret_type"`
	SecretOwner string `json:"secret_owner"`
}

// SecretListRequest соответствует protobuf SecretListRequest
type SecretListFilterDBRequest struct {
	SecretOwner string `json:"secret_owner"`
}
