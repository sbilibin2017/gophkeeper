package client

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/sbilibin2017/gophkeeper/internal/models"
)

// AddBankCardLocal inserts a BankCardAddRequest into the local DB.
func AddBankCardLocal(
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
	return err
}