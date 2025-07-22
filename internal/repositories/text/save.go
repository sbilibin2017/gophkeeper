package text

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/sbilibin2017/gophkeeper/internal/models"
)

// SaveRepository provides methods to save or update text data in the local database.
type SaveRepository struct {
	db *sqlx.DB
}

// NewSaveRepository creates a new SaveRepository with the given database connection.
func NewSaveRepository(db *sqlx.DB) *SaveRepository {
	return &SaveRepository{db: db}
}

// Save inserts or updates a text secret identified by secret_name.
func (r *SaveRepository) Save(ctx context.Context, req *models.TextAddRequest) error {
	query := `
		INSERT INTO text_client (secret_name, content, meta, updated_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (secret_name) DO UPDATE SET
			content = EXCLUDED.content,
			meta = EXCLUDED.meta,
			updated_at = EXCLUDED.updated_at
	`

	updatedAt := time.Now().Format(time.RFC3339)

	_, err := r.db.ExecContext(ctx, query,
		req.SecretName,
		req.Content,
		req.Meta,
		updatedAt,
	)
	return err
}
