package user

import (
	"context"

	"github.com/jmoiron/sqlx"
)

// CreateClientTable creates the user_client table.
// Returns an error if the table already exists.
func CreateClientTable(ctx context.Context, db *sqlx.DB) error {
	query := `
	CREATE TABLE user_client (
		secret_name TEXT PRIMARY KEY,
		secret_owner TEXT NOT NULL,
		username TEXT NOT NULL,
		password TEXT NOT NULL,
		meta TEXT,
		updated_at TEXT NOT NULL
	);
	`
	_, err := db.ExecContext(ctx, query)
	return err
}

// DropClientTable drops the user_client table.
func DropClientTable(ctx context.Context, db *sqlx.DB) error {
	query := `DROP TABLE user_client;`
	_, err := db.ExecContext(ctx, query)
	return err
}
