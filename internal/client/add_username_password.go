package client

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/sbilibin2017/gophkeeper/internal/models"
)

// AddUsernamePasswordLocal inserts or updates a UsernamePasswordAddRequest in the local DB.
func AddUsernamePasswordLocal(
	ctx context.Context,
	db *sqlx.DB,
	req models.UsernamePasswordAddRequest,
) error {
	query := `
		INSERT INTO secret_username_password_request (secret_name, username, password, meta)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (secret_name) DO UPDATE SET
			username = EXCLUDED.username,
			password = EXCLUDED.password,
			meta = EXCLUDED.meta;
	`
	_, err := db.ExecContext(ctx, query, req.SecretName, req.Username, req.Password, req.Meta)
	return err
}
