package repositories

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/sbilibin2017/gophkeeper/internal/models"
)

// EncryptedSecretWriteRepository handles write operations for EncryptedSecret using sqlx.
type EncryptedSecretWriteRepository struct {
	db *sqlx.DB
}

func NewEncryptedSecretWriteRepository(db *sqlx.DB) *EncryptedSecretWriteRepository {
	return &EncryptedSecretWriteRepository{db: db}
}

// Save inserts a new encrypted secret into the database.
func (r *EncryptedSecretWriteRepository) Save(ctx context.Context, secret *models.EncryptedSecret) error {
	query := `
		INSERT INTO encrypted_secrets (secret_name, secret_type, ciphertext, aes_key_enc, timestamp)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (secret_name) DO UPDATE SET
			secret_type = EXCLUDED.secret_type,
			ciphertext = EXCLUDED.ciphertext,
			aes_key_enc = EXCLUDED.aes_key_enc,
			timestamp = EXCLUDED.timestamp;
	`

	_, err := r.db.ExecContext(ctx, query,
		secret.SecretName,
		secret.SecretType,
		secret.Ciphertext,
		secret.AESKeyEnc,
		secret.Timestamp,
	)
	if err != nil {
		return fmt.Errorf("failed to save secret '%s': %w", secret.SecretName, err)
	}
	return nil
}

// EncryptedSecretReadRepository handles read operations for EncryptedSecret using sqlx.
type EncryptedSecretReadRepository struct {
	db *sqlx.DB
}

func NewEncryptedSecretReadRepository(db *sqlx.DB) *EncryptedSecretReadRepository {
	return &EncryptedSecretReadRepository{db: db}
}

// Get retrieves a single encrypted secret by its secret name.
func (r *EncryptedSecretReadRepository) Get(ctx context.Context, secretName string) (*models.EncryptedSecret, error) {
	query := `
		SELECT secret_name, secret_type, ciphertext, aes_key_enc, timestamp
		FROM encrypted_secrets
		WHERE secret_name = $1
	`

	var secret models.EncryptedSecret
	err := r.db.GetContext(ctx, &secret, query, secretName)
	if err != nil {
		return nil, fmt.Errorf("failed to get secret '%s': %w", secretName, err)
	}
	return &secret, nil
}

// List fetches all encrypted secrets.
func (r *EncryptedSecretReadRepository) List(ctx context.Context) ([]*models.EncryptedSecret, error) {
	query := `
		SELECT secret_name, secret_type, ciphertext, aes_key_enc, timestamp
		FROM encrypted_secrets
	`

	var secrets []*models.EncryptedSecret
	err := r.db.SelectContext(ctx, &secrets, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list secrets: %w", err)
	}
	return secrets, nil
}

// CreateEncryptedSecretsTable creates the encrypted_secrets table.
func CreateEncryptedSecretsTable(ctx context.Context, db *sqlx.DB) error {
	const query = `
	CREATE TABLE encrypted_secrets (
		secret_name TEXT PRIMARY KEY,
		secret_type TEXT NOT NULL,
		ciphertext BYTEA NOT NULL,
		aes_key_enc BYTEA NOT NULL,
		timestamp BIGINT NOT NULL
	);
	`
	_, err := db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to create encrypted_secrets table: %w", err)
	}
	return nil
}

// DropEncryptedSecretsTable drops the encrypted_secrets table.
func DropEncryptedSecretsTable(ctx context.Context, db *sqlx.DB) error {
	const query = `DROP TABLE encrypted_secrets;`
	_, err := db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to drop encrypted_secrets table: %w", err)
	}
	return nil
}
