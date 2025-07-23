package repositories

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/sbilibin2017/gophkeeper/internal/models"
)

// TextWriteRepository handles write operations for textual secrets.
type TextWriteRepository struct {
	db *sqlx.DB
}

// NewTextWriteRepository creates a new TextWriteRepository with the given DB connection.
func NewTextWriteRepository(db *sqlx.DB) *TextWriteRepository {
	return &TextWriteRepository{db: db}
}

// Add inserts or updates a text secret in the database.
func (r *TextWriteRepository) Add(ctx context.Context, secret *models.Text) error {
	const query = `
		INSERT INTO texts (secret_name, secret_owner, data, meta, updated_at)
		VALUES (:secret_name, :secret_owner, :data, :meta, :updated_at)
		ON CONFLICT (secret_name) DO UPDATE SET
			secret_owner = EXCLUDED.secret_owner,
			data = EXCLUDED.data,
			meta = EXCLUDED.meta,
			updated_at = EXCLUDED.updated_at
	`

	_, err := r.db.NamedExecContext(ctx, query, secret)
	return err
}

// TextReadRepository handles read-only operations for textual secrets.
type TextReadRepository struct {
	db *sqlx.DB
}

// NewTextReadRepository creates a new TextReadRepository with the given DB connection.
func NewTextReadRepository(db *sqlx.DB) *TextReadRepository {
	return &TextReadRepository{db: db}
}

// List retrieves all text secrets from the database.
func (r *TextReadRepository) List(ctx context.Context) ([]*models.Text, error) {
	const query = `
		SELECT secret_name, secret_owner, data, meta, updated_at
		FROM texts
	`

	var texts []*models.Text
	if err := r.db.SelectContext(ctx, &texts, query); err != nil {
		return nil, err
	}
	return texts, nil
}
