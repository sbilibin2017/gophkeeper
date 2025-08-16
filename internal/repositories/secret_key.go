package repositories

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/sbilibin2017/gophkeeper/internal/models"
)

// SecretKeyWriteRepository управляет записью зашифрованных AES ключей
type SecretKeyWriteRepository struct {
	db *sqlx.DB
}

// NewSecretKeyWriteRepository создаёт новый репозиторий записи AES ключей
func NewSecretKeyWriteRepository(db *sqlx.DB) *SecretKeyWriteRepository {
	return &SecretKeyWriteRepository{db: db}
}

// Save вставляет новый AES ключ или обновляет существующий по SecretKeyID
func (r *SecretKeyWriteRepository) Save(
	ctx context.Context,
	secretKeyID, secretID, deviceID string,
	encryptedAESKey []byte,
) error {
	query := `
	INSERT INTO secret_keys (secret_key_id, secret_id, device_id, encrypted_aes_key, created_at, updated_at)
	VALUES ($1, $2, $3, $4, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	ON CONFLICT(secret_key_id) DO UPDATE SET
		secret_id = EXCLUDED.secret_id,
		device_id = EXCLUDED.device_id,
		encrypted_aes_key = EXCLUDED.encrypted_aes_key,
		updated_at = CURRENT_TIMESTAMP
	`
	_, err := r.db.ExecContext(ctx, query, secretKeyID, secretID, deviceID, encryptedAESKey)
	return err
}

// SecretKeyReadRepository управляет чтением зашифрованных AES ключей
type SecretKeyReadRepository struct {
	db *sqlx.DB
}

// NewSecretKeyReadRepository создаёт новый репозиторий чтения AES ключей
func NewSecretKeyReadRepository(db *sqlx.DB) *SecretKeyReadRepository {
	return &SecretKeyReadRepository{db: db}
}

// Get возвращает запись AES ключа по secretID и deviceID
func (r *SecretKeyReadRepository) Get(
	ctx context.Context,
	secretID, deviceID string,
) (*models.SecretKeyDB, error) {
	var secretKey models.SecretKeyDB
	query := `
	SELECT secret_key_id, secret_id, device_id, encrypted_aes_key, created_at, updated_at
	FROM secret_keys
	WHERE secret_id = $1 AND device_id = $2
	`
	if err := r.db.GetContext(ctx, &secretKey, query, secretID, deviceID); err != nil {
		return nil, err
	}
	return &secretKey, nil
}
