package repositories

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/sbilibin2017/gophkeeper/internal/models"
)

// SecretWriteRepository handles write operations related to secrets.
type SecretWriteRepository struct {
	db *sqlx.DB
}

func NewSecretWriteRepository(db *sqlx.DB) *SecretWriteRepository {
	return &SecretWriteRepository{db: db}
}

// Save inserts or updates a secret, taking explicit arguments.
func (r *SecretWriteRepository) Save(
	ctx context.Context,
	secretOwner string,
	secretName string,
	secretType string,
	ciphertext []byte,
	aesKeyEnc []byte,
) error {
	query := `
		INSERT INTO secrets (secret_name, secret_type, secret_owner, ciphertext, aes_key_enc, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		ON CONFLICT(secret_name, secret_type, secret_owner) DO UPDATE SET
			ciphertext = EXCLUDED.ciphertext,
			aes_key_enc = EXCLUDED.aes_key_enc,
			updated_at = CURRENT_TIMESTAMP;
	`
	_, err := r.db.ExecContext(ctx, query,
		secretName,
		secretType,
		secretOwner,
		ciphertext,
		aesKeyEnc,
	)
	if err != nil {
		return fmt.Errorf("failed to save secret: %w", err)
	}
	return nil
}

// SecretReadRepository handles read operations related to secrets.
type SecretReadRepository struct {
	db *sqlx.DB
}

func NewSecretReadRepository(db *sqlx.DB) *SecretReadRepository {
	return &SecretReadRepository{db: db}
}

// Get fetches a secret by name, type, and owner.
func (r *SecretReadRepository) Get(
	ctx context.Context,
	secretOwner string,
	secretType string,
	secretName string,
) (*models.Secret, error) {
	query := `
		SELECT secret_name, secret_type, secret_owner, ciphertext, aes_key_enc, created_at, updated_at
		FROM secrets
		WHERE secret_name = $1 AND secret_type = $2 AND secret_owner = $3
	`

	var secret models.Secret
	err := r.db.GetContext(ctx, &secret, query,
		secretName,
		secretType,
		secretOwner,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get secret: %w", err)
	}
	return &secret, nil
}

// List fetches all secrets for a given owner.
func (r *SecretReadRepository) List(
	ctx context.Context,
	secretOwner string,
) ([]*models.Secret, error) {
	query := `
		SELECT secret_name, secret_type, secret_owner, ciphertext, aes_key_enc, created_at, updated_at
		FROM secrets
		WHERE secret_owner = $1
	`

	var secrets []*models.Secret
	err := r.db.SelectContext(ctx, &secrets, query, secretOwner)
	if err != nil {
		return nil, fmt.Errorf("failed to list secrets: %w", err)
	}
	return secrets, nil
}
