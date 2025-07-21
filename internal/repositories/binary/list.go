package binary

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/sbilibin2017/gophkeeper/internal/models"
)

// ListRepository provides methods to list binary data from the local database.
type ListRepository struct {
	db *sqlx.DB
}

// NewListRepository creates a new ListRepository with the given database connection.
func NewListRepository(db *sqlx.DB) *ListRepository {
	return &ListRepository{db: db}
}

// List retrieves all binary secrets stored in the local database.
func (r *ListRepository) List(ctx context.Context) ([]models.BinaryAddRequest, error) {
	query := `
		SELECT secret_name, data, meta
		FROM binary_client
	`

	var binaries []models.BinaryAddRequest
	err := r.db.SelectContext(ctx, &binaries, query)
	if err != nil {
		return nil, err
	}

	return binaries, nil
}
