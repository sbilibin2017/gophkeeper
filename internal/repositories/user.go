package repositories

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/sbilibin2017/gophkeeper/internal/models"
)

// UserWriteRepository handles write operations for user secrets.
type UserWriteRepository struct {
	db *sqlx.DB
}

// NewUserWriteRepository creates a new UserWriteRepository with the given DB connection.
func NewUserWriteRepository(db *sqlx.DB) *UserWriteRepository {
	return &UserWriteRepository{db: db}
}

// Add inserts or updates a user secret in the database.
func (r *UserWriteRepository) Add(ctx context.Context, secret *models.User) error {
	const query = `
		INSERT INTO users (secret_name, secret_owner, login, password, meta, updated_at)
		VALUES (:secret_name, :secret_owner, :login, :password, :meta, :updated_at)
		ON CONFLICT (secret_name) DO UPDATE SET
			secret_owner = EXCLUDED.secret_owner,
			login = EXCLUDED.login,
			password = EXCLUDED.password,
			meta = EXCLUDED.meta,
			updated_at = EXCLUDED.updated_at
	`

	_, err := r.db.NamedExecContext(ctx, query, secret)
	return err
}

// UserReadRepository handles read-only operations for user secrets.
type UserReadRepository struct {
	db *sqlx.DB
}

// NewUserReadRepository creates a new UserReadRepository with the given DB connection.
func NewUserReadRepository(db *sqlx.DB) *UserReadRepository {
	return &UserReadRepository{db: db}
}

// List retrieves all user secrets from the database.
func (r *UserReadRepository) List(ctx context.Context) ([]*models.User, error) {
	const query = `
		SELECT secret_name, secret_owner, login, password, meta, updated_at
		FROM users
	`

	var users []*models.User
	if err := r.db.SelectContext(ctx, &users, query); err != nil {
		return nil, err
	}
	return users, nil
}
