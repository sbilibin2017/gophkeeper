package bankcard

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/sbilibin2017/gophkeeper/internal/models"
)

// ListRepository provides methods to list bank card data from the local database.
type ListRepository struct {
	db *sqlx.DB
}

// NewListRepository creates a new ListRepository with the given database connection.
func NewListRepository(db *sqlx.DB) *ListRepository {
	return &ListRepository{db: db}
}

// List retrieves all bank cards stored in the local database.
func (r *ListRepository) List(ctx context.Context) ([]models.BankCardAddRequest, error) {
	query := `
		SELECT secret_name, number, owner, exp, cvv, meta
		FROM bankcard_client
	`

	var bankcards []models.BankCardAddRequest
	err := r.db.SelectContext(ctx, &bankcards, query)
	if err != nil {
		return nil, err
	}

	return bankcards, nil
}
