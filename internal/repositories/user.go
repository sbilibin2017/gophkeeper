package repositories

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/sbilibin2017/gophkeeper/internal/models"
)

// UserWriteRepository handles write operations for users.
type UserWriteRepository struct {
	db *sqlx.DB
}

func NewUserWriteRepository(db *sqlx.DB) *UserWriteRepository {
	return &UserWriteRepository{db: db}
}

// Save inserts or updates a user record.
func (r *UserWriteRepository) Save(
	ctx context.Context,
	username string,
	passwordHash string,
) error {
	query := `
		INSERT INTO users (username, password_hash, created_at, updated_at)
		VALUES ($1, $2, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		ON CONFLICT(username) DO UPDATE SET
			password_hash = EXCLUDED.password_hash,
			updated_at = CURRENT_TIMESTAMP;
	`
	_, err := r.db.ExecContext(ctx, query, username, passwordHash)
	if err != nil {
		return fmt.Errorf("failed to save user: %w", err)
	}
	return nil
}

// UserReadRepository handles read operations for users.
type UserReadRepository struct {
	db *sqlx.DB
}

func NewUserReadRepository(db *sqlx.DB) *UserReadRepository {
	return &UserReadRepository{db: db}
}

// Get fetches a user by username.
func (r *UserReadRepository) Get(
	ctx context.Context,
	username string,
) (*models.UserDB, error) {
	query := `
		SELECT username, password_hash, created_at, updated_at
		FROM users
		WHERE username = $1;
	`
	var user models.UserDB
	err := r.db.GetContext(ctx, &user, query, username)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return &user, nil
}
