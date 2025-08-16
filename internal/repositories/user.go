package repositories

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/sbilibin2017/gophkeeper/internal/models"
)

// UserWriteRepository управляет записью пользователей
type UserWriteRepository struct {
	db *sqlx.DB
}

// NewUserWriteRepository создаёт новый репозиторий записи пользователей
func NewUserWriteRepository(db *sqlx.DB) *UserWriteRepository {
	return &UserWriteRepository{db: db}
}

// Save вставляет нового пользователя или обновляет существующего по user_id
func (r *UserWriteRepository) Save(
	ctx context.Context,
	userID, username, passwordHash string,
) error {
	query := `
	INSERT INTO users (user_id, username, password_hash, created_at, updated_at)
	VALUES ($1, $2, $3, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	ON CONFLICT(user_id) DO UPDATE SET
		username = EXCLUDED.username,
		password_hash = EXCLUDED.password_hash,
		updated_at = CURRENT_TIMESTAMP
	`
	_, err := r.db.ExecContext(ctx, query, userID, username, passwordHash)
	return err
}

// UserReadRepository управляет чтением пользователей
type UserReadRepository struct {
	db *sqlx.DB
}

// NewUserReadRepository создаёт новый репозиторий чтения пользователей
func NewUserReadRepository(db *sqlx.DB) *UserReadRepository {
	return &UserReadRepository{db: db}
}

// Get возвращает пользователя по username
func (r *UserReadRepository) Get(
	ctx context.Context,
	username string,
) (*models.UserDB, error) {
	var user models.UserDB
	query := `SELECT user_id, username, password_hash, created_at, updated_at FROM users WHERE username = $1`
	if err := r.db.GetContext(ctx, &user, query, username); err != nil {
		return nil, err
	}
	return &user, nil
}
