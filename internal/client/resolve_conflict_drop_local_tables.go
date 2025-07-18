package client

import (
	"context"

	"github.com/jmoiron/sqlx"
)

// DropBinaryRequestTable drops the "secret_binary_request" table from the database.
func DropBinaryRequestTable(ctx context.Context, db *sqlx.DB) error {
	_, err := db.ExecContext(ctx, `DROP TABLE IF EXISTS secret_binary_request;`)
	return err
}

// DropTextRequestTable drops the "secret_text_request" table from the database.
func DropTextRequestTable(ctx context.Context, db *sqlx.DB) error {
	_, err := db.ExecContext(ctx, `DROP TABLE IF EXISTS secret_text_request;`)
	return err
}

// DropUsernamePasswordRequestTable drops the "secret_username_password_request" table from the database.
func DropUsernamePasswordRequestTable(ctx context.Context, db *sqlx.DB) error {
	_, err := db.ExecContext(ctx, `DROP TABLE IF EXISTS secret_username_password_request;`)
	return err
}

// DropBankCardRequestTable drops the "secret_bank_card_request" table from the database.
func DropBankCardRequestTable(ctx context.Context, db *sqlx.DB) error {
	_, err := db.ExecContext(ctx, `DROP TABLE IF EXISTS secret_bank_card_request;`)
	return err
}
