package repositories

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/sbilibin2017/gophkeeper/internal/models"
)

// BankCardWriteRepository implements BankCardAdder backed by SQL database.
type BankCardWriteRepository struct {
	db *sqlx.DB
}

// NewBankCardWriteRepository creates a new BankCardWriteRepository with given DB connection.
func NewBankCardWriteRepository(db *sqlx.DB) *BankCardWriteRepository {
	return &BankCardWriteRepository{db: db}
}

// Add inserts or updates a bank card secret in the database.
func (r *BankCardWriteRepository) Add(ctx context.Context, secret *models.BankCard) error {
	const query = `
		INSERT INTO bankcards (secret_name, secret_owner, number, owner, exp, cvv, meta, updated_at)
		VALUES (:secret_name, :secret_owner, :number, :owner, :exp, :cvv, :meta, :updated_at)
		ON CONFLICT (secret_name) DO UPDATE SET
			secret_owner = EXCLUDED.secret_owner,
			number = EXCLUDED.number,
			owner = EXCLUDED.owner,
			exp = EXCLUDED.exp,
			cvv = EXCLUDED.cvv,
			meta = EXCLUDED.meta,
			updated_at = EXCLUDED.updated_at
	`

	_, err := r.db.NamedExecContext(ctx, query, secret)
	return err
}

// BankCardReadRepository handles read-only operations for bank cards.
type BankCardReadRepository struct {
	db *sqlx.DB
}

// NewBankCardReadRepository creates a new BankCardReadRepository with the given DB connection.
func NewBankCardReadRepository(db *sqlx.DB) *BankCardReadRepository {
	return &BankCardReadRepository{db: db}
}

// List retrieves all bank card secrets from the database.
func (r *BankCardReadRepository) List(ctx context.Context) ([]*models.BankCard, error) {
	const query = `
		SELECT secret_name, secret_owner, number, owner, exp, cvv, meta, updated_at 
		FROM bankcards
	`

	var bankCards []*models.BankCard
	if err := r.db.SelectContext(ctx, &bankCards, query); err != nil {
		return nil, err
	}
	return bankCards, nil
}
