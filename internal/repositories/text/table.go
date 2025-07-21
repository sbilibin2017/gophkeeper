package text

import (
	"context"

	"github.com/jmoiron/sqlx"
)

// CreateClientTable creates the text_client table.
// Returns an error if the table already exists.
func CreateClientTable(ctx context.Context, db *sqlx.DB) error {
	query := `
	CREATE TABLE text_client (
		secret_name TEXT PRIMARY KEY,		
		content TEXT NOT NULL,
		meta TEXT,
		updated_at TEXT NOT NULL
	);
	`
	_, err := db.ExecContext(ctx, query)
	return err
}

// DropClientTable drops the text_client table.
func DropClientTable(ctx context.Context, db *sqlx.DB) error {
	query := `DROP TABLE text_client;`
	_, err := db.ExecContext(ctx, query)
	return err
}
