package binary

import (
	"context"

	"github.com/jmoiron/sqlx"
)

// CreateClientTable creates the binary_client table.
// Returns error if the table already exists.
func CreateClientTable(ctx context.Context, db *sqlx.DB) error {
	query := `
	CREATE TABLE binary_client (
		secret_name TEXT PRIMARY KEY,
		secret_owner TEXT NOT NULL,
		data BLOB NOT NULL,
		meta TEXT,
		updated_at TEXT NOT NULL
	);
	`
	_, err := db.ExecContext(ctx, query)
	return err
}

// DropClientTable drops the binary_client table.
func DropClientTable(ctx context.Context, db *sqlx.DB) error {
	query := `DROP TABLE binary_client;`
	_, err := db.ExecContext(ctx, query)
	return err
}
