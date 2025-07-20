package bankcard

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/sbilibin2017/gophkeeper/internal/client/models"
)

// SaveRepository provides methods to save or update bank card data in the local database.
type SaveRepository struct {
	db *sqlx.DB
}

// NewSaveRepository creates a new SaveRepository with the given database connection.
func NewSaveRepository(db *sqlx.DB) *SaveRepository {
	return &SaveRepository{db: db}
}

// Save inserts a new bank card or updates an existing one based on the secret name.
// It uses an UPSERT query to ensure idempotency.
func (r *SaveRepository) Save(ctx context.Context, req *models.BankCardAddRequest) error {
	query := `
		INSERT INTO bankcard_client (secret_name, number, owner, exp, cvv, meta)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (secret_name) DO UPDATE SET
			number = EXCLUDED.number,
			owner = EXCLUDED.owner,
			exp = EXCLUDED.exp,
			cvv = EXCLUDED.cvv,
			meta = EXCLUDED.meta
	`
	_, err := r.db.ExecContext(ctx, query,
		req.SecretName,
		req.Number,
		req.Owner,
		req.Exp,
		req.CVV,
		req.Meta,
	)
	return err
}
