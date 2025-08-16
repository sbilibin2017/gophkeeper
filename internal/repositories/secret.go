package repositories

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/sbilibin2017/gophkeeper/internal/models"
)

// SecretWriteRepository управляет записью секретов пользователей
type SecretWriteRepository struct {
	db *sqlx.DB
}

// NewSecretWriteRepository создаёт новый репозиторий записи секретов
func NewSecretWriteRepository(db *sqlx.DB) *SecretWriteRepository {
	return &SecretWriteRepository{db: db}
}

// Save вставляет новый секрет или обновляет существующий по secret_id
func (r *SecretWriteRepository) Save(
	ctx context.Context,
	secretID, userID, secretName, secretType string,
	encryptedPayload, nonce []byte,
	meta string,
) error {
	query := `
	INSERT INTO secrets (secret_id, user_id, secret_name, secret_type, encrypted_payload, nonce, meta, created_at, updated_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	ON CONFLICT(secret_id) DO UPDATE SET
		user_id = EXCLUDED.user_id,
		secret_name = EXCLUDED.secret_name,
		secret_type = EXCLUDED.secret_type,
		encrypted_payload = EXCLUDED.encrypted_payload,
		nonce = EXCLUDED.nonce,
		meta = EXCLUDED.meta,
		updated_at = CURRENT_TIMESTAMP
	`
	_, err := r.db.ExecContext(ctx, query, secretID, userID, secretName, secretType, encryptedPayload, nonce, meta)
	return err
}

// SecretReadRepository управляет чтением секретов пользователей
type SecretReadRepository struct {
	db *sqlx.DB
}

// NewSecretReadRepository создаёт новый репозиторий чтения секретов
func NewSecretReadRepository(db *sqlx.DB) *SecretReadRepository {
	return &SecretReadRepository{db: db}
}

// Get возвращает секрет по userID и secretName
func (r *SecretReadRepository) Get(
	ctx context.Context,
	userID, secretName string,
) (*models.SecretDB, error) {
	var secret models.SecretDB
	query := `
	SELECT secret_id, user_id, secret_name, secret_type, encrypted_payload, nonce, meta, created_at, updated_at
	FROM secrets
	WHERE user_id = $1 AND secret_name = $2
	`
	if err := r.db.GetContext(ctx, &secret, query, userID, secretName); err != nil {
		return nil, err
	}
	return &secret, nil
}

// List возвращает все секреты пользователя по userID
func (r *SecretReadRepository) List(
	ctx context.Context,
	userID string,
) ([]*models.SecretDB, error) {
	var secrets []*models.SecretDB
	query := `
	SELECT secret_id, user_id, secret_name, secret_type, encrypted_payload, nonce, meta, created_at, updated_at
	FROM secrets
	WHERE user_id = $1
	ORDER BY created_at DESC
	`
	if err := r.db.SelectContext(ctx, &secrets, query, userID); err != nil {
		return nil, err
	}
	return secrets, nil
}
