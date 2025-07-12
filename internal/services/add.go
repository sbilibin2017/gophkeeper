package services

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/sbilibin2017/gophkeeper/internal/models"
)

func AddLoginPassword(
	ctx context.Context,
	db *sql.DB,
	secret *models.UsernamePassword,
) error {
	metaJSON, err := json.Marshal(secret.Meta)
	if err != nil {
		return err
	}

	query := `
	INSERT INTO username_passwords (username, password, meta) 
	VALUES ($1, $2, $3)
	`

	_, err = db.ExecContext(ctx, query, secret.Username, secret.Password, metaJSON)
	if err != nil {
		return err
	}

	return nil
}

func AddText(
	ctx context.Context,
	db *sql.DB,
	text *models.Text,
) error {
	metaJSON, err := json.Marshal(text.Meta)
	if err != nil {
		return err
	}

	query := `
	INSERT INTO texts (content, meta)
	VALUES ($1, $2)
	`

	_, err = db.ExecContext(ctx, query, text.Content, metaJSON)
	if err != nil {
		return err
	}

	return nil
}

func AddBinary(
	ctx context.Context,
	db *sql.DB,
	bin *models.Binary,
) error {
	metaJSON, err := json.Marshal(bin.Meta)
	if err != nil {
		return err
	}

	query := `
	INSERT INTO binaries (data, meta)
	VALUES ($1, $2)
	`

	_, err = db.ExecContext(ctx, query, bin.Data, metaJSON)
	if err != nil {
		return err
	}

	return nil
}

func AddBankCard(
	ctx context.Context,
	db *sql.DB,
	card *models.BankCard,
) error {
	metaJSON, err := json.Marshal(card.Meta)
	if err != nil {
		return err
	}

	query := `
	INSERT INTO bank_cards (number, owner, expiry, cvv, meta)
	VALUES ($1, $2, $3, $4, $5)
	`

	_, err = db.ExecContext(ctx, query, card.Number, card.Owner, card.Expiry, card.CVV, metaJSON)
	if err != nil {
		return err
	}

	return nil
}
