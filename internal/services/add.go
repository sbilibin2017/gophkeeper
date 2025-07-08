package services

import (
	"context"
	"encoding/base64"
	"encoding/json"

	"github.com/jmoiron/sqlx"
	"github.com/sbilibin2017/gophkeeper/internal/models"
)

// AddLoginPassword saves or updates a LoginPassword secret in the database.
// On conflict by secret_id, updates login, password, and metadata.
func AddLoginPassword(ctx context.Context, db *sqlx.DB, secret *models.LoginPassword) error {
	metaJSON, err := json.Marshal(secret.Meta)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO login_passwords (secret_id, login, password, meta)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (secret_id) DO UPDATE SET
			login = EXCLUDED.login,
			password = EXCLUDED.password,
			meta = EXCLUDED.meta
	`

	_, err = db.ExecContext(ctx, query, secret.SecretID, secret.Login, secret.Password, string(metaJSON))
	return err
}

// AddText saves or updates a text secret in the database.
// On conflict by secret_id, updates content, metadata, and updated_at.
func AddText(ctx context.Context, db *sqlx.DB, secret *models.Text) error {
	metaJSON, err := json.Marshal(secret.Meta)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO texts (secret_id, content, meta, updated_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (secret_id) DO UPDATE SET
			content = EXCLUDED.content,
			meta = EXCLUDED.meta,
			updated_at = EXCLUDED.updated_at
	`

	_, err = db.ExecContext(ctx, query, secret.SecretID, secret.Content, string(metaJSON), secret.UpdatedAt)
	return err
}

// AddCard saves or updates a card secret in the database.
// On conflict by secret_id, updates card details, metadata, and updated_at.
func AddCard(ctx context.Context, db *sqlx.DB, secret *models.Card) error {
	metaJSON, err := json.Marshal(secret.Meta)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO cards (secret_id, number, holder, exp_month, exp_year, cvv, meta, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (secret_id) DO UPDATE SET
			number = EXCLUDED.number,
			holder = EXCLUDED.holder,
			exp_month = EXCLUDED.exp_month,
			exp_year = EXCLUDED.exp_year,
			cvv = EXCLUDED.cvv,
			meta = EXCLUDED.meta,
			updated_at = EXCLUDED.updated_at
	`

	_, err = db.ExecContext(ctx, query,
		secret.SecretID,
		secret.Number,
		secret.Holder,
		secret.ExpMonth,
		secret.ExpYear,
		secret.CVV,
		string(metaJSON),
		secret.UpdatedAt,
	)
	return err
}

// AddBinary saves or updates a binary secret in the database.
// The data is base64-encoded before saving.
// On conflict by secret_id, updates data, metadata, and updated_at.
func AddBinary(ctx context.Context, db *sqlx.DB, secret *models.Binary) error {
	metaJSON, err := json.Marshal(secret.Meta)
	if err != nil {
		return err
	}

	base64Data := base64.StdEncoding.EncodeToString(secret.Data)

	query := `
		INSERT INTO binaries (secret_id, data, meta, updated_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (secret_id) DO UPDATE SET
			data = EXCLUDED.data,
			meta = EXCLUDED.meta,
			updated_at = EXCLUDED.updated_at
	`

	_, err = db.ExecContext(ctx, query,
		secret.SecretID,
		base64Data,
		string(metaJSON),
		secret.UpdatedAt,
	)
	return err
}
