package client

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/sbilibin2017/gophkeeper/internal/models"
)

// AddTextLocal inserts a TextAddRequest into the local DB.
func AddTextLocal(
	ctx context.Context,
	db *sqlx.DB,
	req models.TextAddRequest,
) error {
	query := `
		INSERT INTO secret_text_request (secret_name, content, meta)
		VALUES ($1, $2, $3)
		ON CONFLICT (secret_name) DO UPDATE SET
			content = EXCLUDED.content,
			meta = EXCLUDED.meta;
	`
	_, err := db.ExecContext(ctx, query, req.SecretName, req.Content, req.Meta)
	return err
}
