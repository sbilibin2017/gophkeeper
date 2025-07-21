package binary

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/sbilibin2017/gophkeeper/internal/models"
)

// SaveRepository provides methods to save or update binary data in the local database.
type SaveRepository struct {
	db *sqlx.DB
}

// NewSaveRepository creates a new SaveRepository with the given database connection.
func NewSaveRepository(db *sqlx.DB) *SaveRepository {
	return &SaveRepository{db: db}
}

// Save inserts or updates a binary secret identified by secret_name.
func (r *SaveRepository) Save(ctx context.Context, req *models.BinaryAddRequest) error {
	query := `
		INSERT INTO binary_client (secret_name, data, meta, updated_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (secret_name) DO UPDATE SET
			data = EXCLUDED.data,
			meta = EXCLUDED.meta,
			updated_at = EXCLUDED.updated_at
	`

	updatedAt := time.Now().Format(time.RFC3339)

	_, err := r.db.ExecContext(ctx, query,
		req.SecretName,
		req.Data,
		req.Meta,
		updatedAt,
	)
	return err
}
