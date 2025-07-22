package user

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/sbilibin2017/gophkeeper/internal/models"
)

// SaveRepository provides methods to save or update user data in the local database.
type SaveRepository struct {
	db *sqlx.DB
}

// NewSaveRepository creates a new SaveRepository with the given database connection.
func NewSaveRepository(db *sqlx.DB) *SaveRepository {
	return &SaveRepository{db: db}
}

// Save inserts a new user secret or updates an existing one based on the secret_name.
// It uses an UPSERT query to ensure idempotency.
func (r *SaveRepository) Save(ctx context.Context, req *models.UserAddRequest) error {
	query := `
		INSERT INTO user_client (secret_name, username, password, meta, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (secret_name) DO UPDATE SET
			username = EXCLUDED.username,
			password = EXCLUDED.password,
			meta = EXCLUDED.meta,
			updated_at = EXCLUDED.updated_at
	`

	updatedAt := time.Now().Format(time.RFC3339)

	_, err := r.db.ExecContext(ctx, query,
		req.SecretName,
		req.Username,
		req.Password,
		req.Meta,
		updatedAt,
	)
	return err
}
