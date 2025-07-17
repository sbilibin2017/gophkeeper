package client

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/sbilibin2017/gophkeeper/internal/models"
)

// AddBinaryLocal inserts a BinaryAddRequest into the local DB.
func AddBinaryLocal(
	ctx context.Context,
	db *sqlx.DB,
	req models.BinaryAddRequest,
) error {
	query := `
		INSERT INTO secret_binary_request (secret_name, data, meta)
		VALUES ($1, $2, $3)
		ON CONFLICT (secret_name) DO UPDATE SET
			data = EXCLUDED.data,
			meta = EXCLUDED.meta;
	`
	_, err := db.ExecContext(ctx, query, req.SecretName, req.Data, req.Meta)
	return err
}
