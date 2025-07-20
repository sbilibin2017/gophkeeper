package bankcard

import (
	"context"

	"github.com/jmoiron/sqlx"
)

func CreateClientTable(ctx context.Context, db *sqlx.DB) error {
	query := `
	CREATE TABLE bankcard_client (
		secret_name TEXT PRIMARY KEY,
		number TEXT NOT NULL,
		owner TEXT NOT NULL,
		exp TEXT NOT NULL,
		cvv TEXT NOT NULL,
		meta TEXT
	);
	`
	_, err := db.ExecContext(ctx, query)
	return err
}

// DropClientTable drops the bankcard_client table.
func DropClientTable(ctx context.Context, db *sqlx.DB) error {
	query := `DROP TABLE bankcard_client;`
	_, err := db.ExecContext(ctx, query)
	return err
}
