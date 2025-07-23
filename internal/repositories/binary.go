package repositories

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/sbilibin2017/gophkeeper/internal/models"
)

// BinaryWriteRepository handles write operations for binary secrets.
type BinaryWriteRepository struct {
	db *sqlx.DB
}

// NewBinaryWriteRepository creates a new BinaryWriteRepository with the given DB connection.
func NewBinaryWriteRepository(db *sqlx.DB) *BinaryWriteRepository {
	return &BinaryWriteRepository{db: db}
}

// Add inserts or updates a binary secret in the database.
func (r *BinaryWriteRepository) Add(ctx context.Context, secret *models.Binary) error {
	const query = `
		INSERT INTO binaries (secret_name, secret_owner, file_path, data, meta, updated_at)
		VALUES (:secret_name, :secret_owner, :file_path, :data, :meta, :updated_at)
		ON CONFLICT (secret_name) DO UPDATE SET
			secret_owner = EXCLUDED.secret_owner,
			file_path = EXCLUDED.file_path,
			data = EXCLUDED.data,
			meta = EXCLUDED.meta,
			updated_at = EXCLUDED.updated_at
	`

	_, err := r.db.NamedExecContext(ctx, query, secret)
	return err
}

// BinaryReadRepository handles read-only operations for binary secrets.
type BinaryReadRepository struct {
	db *sqlx.DB
}

// NewBinaryReadRepository creates a new BinaryReadRepository with the given DB connection.
func NewBinaryReadRepository(db *sqlx.DB) *BinaryReadRepository {
	return &BinaryReadRepository{db: db}
}

// List retrieves all binary secrets from the database.
func (r *BinaryReadRepository) List(ctx context.Context) ([]*models.Binary, error) {
	const query = `
		SELECT secret_name, secret_owner, file_path, data, meta, updated_at
		FROM binaries
	`

	var binaries []*models.Binary
	if err := r.db.SelectContext(ctx, &binaries, query); err != nil {
		return nil, err
	}
	return binaries, nil
}
