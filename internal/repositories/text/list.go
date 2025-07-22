package text

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/sbilibin2017/gophkeeper/internal/models"
)

// ListRepository provides methods to list text data from the local database.
type ListRepository struct {
	db *sqlx.DB
}

// NewListRepository creates a new ListRepository with the given database connection.
func NewListRepository(db *sqlx.DB) *ListRepository {
	return &ListRepository{db: db}
}

// List retrieves all text secrets stored in the local database.
func (r *ListRepository) List(ctx context.Context) ([]models.TextAddRequest, error) {
	query := `
		SELECT secret_name, content, meta
		FROM text_client
	`

	var texts []models.TextAddRequest
	err := r.db.SelectContext(ctx, &texts, query)
	if err != nil {
		return nil, err
	}

	return texts, nil
}
