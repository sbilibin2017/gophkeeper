package client

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/sbilibin2017/gophkeeper/internal/models"
)

// AddBankCardSecret inserts a BankCardAddRequest into the DB.
func AddBankCardSecret(
	ctx context.Context,
	db *sqlx.DB,
	req models.BankCardAddRequest,
) error {
	query := `
		INSERT INTO secret_bank_card_request (secret_name, number, owner, exp, cvv, meta)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (secret_name) DO UPDATE SET
			number = EXCLUDED.number,
			owner = EXCLUDED.owner,
			exp = EXCLUDED.exp,
			cvv = EXCLUDED.cvv,
			meta = EXCLUDED.meta;
	`
	_, err := db.ExecContext(ctx, query, req.SecretName, req.Number, req.Owner, req.Exp, req.CVV, req.Meta)
	if err != nil {
		return err
	}
	return nil
}

// AddUsernamePasswordSecret inserts a UsernamePasswordAddRequest into the DB.
func AddUsernamePasswordSecret(
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
	if err != nil {
		return err
	}
	return nil
}

// AddTextSecret inserts a TextAddRequest into the DB.
func AddTextSecret(
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
	if err != nil {
		return err
	}
	return nil
}

// AddBinarySecret inserts an AddSecretBinaryRequest into the DB.
func AddBinarySecret(
	ctx context.Context,
	db *sqlx.DB,
	req models.AddSecretBinaryRequest,
) error {
	query := `
		INSERT INTO secret_binary_request (secret_name, data, meta)
		VALUES ($1, $2, $3)
		ON CONFLICT (secret_name) DO UPDATE SET
			data = EXCLUDED.data,
			meta = EXCLUDED.meta;
	`
	_, err := db.ExecContext(ctx, query, req.SecretName, req.Data, req.Meta)
	if err != nil {
		return err
	}
	return nil
}
