package user

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/sbilibin2017/gophkeeper/internal/models"
)

// ListRepository provides methods to list user data from the local database.
type ListRepository struct {
	db *sqlx.DB
}

// NewListRepository creates a new ListRepository with the given database connection.
func NewListRepository(db *sqlx.DB) *ListRepository {
	return &ListRepository{db: db}
}

// List retrieves all user secrets stored in the local database.
func (r *ListRepository) List(ctx context.Context) ([]models.UserAddRequest, error) {
	query := `
		SELECT secret_name, username, password, meta
		FROM user_client
	`

	var users []models.UserAddRequest
	err := r.db.SelectContext(ctx, &users, query)
	if err != nil {
		return nil, err
	}

	return users, nil
}
