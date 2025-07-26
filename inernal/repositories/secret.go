package repositories

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/sbilibin2017/gophkeeper/inernal/models"
)

type SecretWriteRepository struct {
	db *sqlx.DB
}

func NewSecretWriteRepository(db *sqlx.DB) *SecretWriteRepository {
	return &SecretWriteRepository{db: db}
}

func (r *SecretWriteRepository) Save(ctx context.Context, secret *models.SecretDB) error {
	query := `
		INSERT INTO secrets (secret_name, secret_type, secret_owner, ciphertext, aes_key_enc, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		ON CONFLICT(secret_name, secret_type, secret_owner) DO UPDATE SET
			ciphertext = EXCLUDED.ciphertext,
			aes_key_enc = EXCLUDED.aes_key_enc,
			updated_at = CURRENT_TIMESTAMP;
	`
	_, err := r.db.ExecContext(ctx, query,
		secret.SecretName,
		secret.SecretType,
		secret.SecretOwner,
		secret.Ciphertext,
		secret.AESKeyEnc,
	)
	if err != nil {
		return err
	}
	return nil
}

type SecretReadRepository struct {
	db *sqlx.DB
}

func NewSecretReadRepository(db *sqlx.DB) *SecretReadRepository {
	return &SecretReadRepository{db: db}
}

// Get using SecretGetFilterDB as filter struct
func (r *SecretReadRepository) Get(ctx context.Context, filter *models.SecretGetFilterDB) (*models.SecretDB, error) {
	query := `
		SELECT secret_name, secret_type, secret_owner, ciphertext, aes_key_enc, created_at, updated_at
		FROM secrets
		WHERE secret_name = $1 AND secret_type = $2 AND secret_owner = $3
	`

	var secret models.SecretDB
	err := r.db.GetContext(ctx, &secret, query,
		filter.SecretName,
		filter.SecretType,
		filter.SecretOwner,
	)
	if err != nil {
		return nil, err
	}
	return &secret, nil
}

// List using SecretListFilterDB as filter struct
func (r *SecretReadRepository) List(ctx context.Context, filter *models.SecretListFilterDB) ([]*models.SecretDB, error) {
	query := `
		SELECT secret_name, secret_type, secret_owner, ciphertext, aes_key_enc, created_at, updated_at
		FROM secrets
		WHERE secret_owner = $1
	`

	var secrets []*models.SecretDB
	err := r.db.SelectContext(ctx, &secrets, query, filter.SecretOwner)
	if err != nil {
		return nil, err
	}
	return secrets, nil
}
