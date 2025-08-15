package repositories

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/sbilibin2017/gophkeeper/internal/models"
)

// UserReaderRepository предоставляет методы чтения данных пользователя из БД.
type UserReaderRepository struct {
	db *sqlx.DB
}

// NewUserReaderRepository создаёт новый экземпляр UserReaderRepository.
func NewUserReaderRepository(db *sqlx.DB) *UserReaderRepository {
	return &UserReaderRepository{db: db}
}

// GetByUsername возвращает пользователя по username.
func (r *UserReaderRepository) GetByUsername(
	ctx context.Context,
	username string,
) (*models.UserDB, error) {
	var query = `
	SELECT user_id, username, password_hash, created_at, updated_at 
	FROM users WHERE username=$1
	`
	var user models.UserDB
	err := r.db.GetContext(ctx, &user, query, username)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// UserWriterRepository предоставляет методы записи данных пользователя в БД.
type UserWriterRepository struct {
	db *sqlx.DB
}

// NewUserWriterRepository создаёт новый экземпляр UserWriterRepository.
func NewUserWriterRepository(db *sqlx.DB) *UserWriterRepository {
	return &UserWriterRepository{db: db}
}

// Save создаёт нового пользователя или обновляет существующего.
func (r *UserWriterRepository) Save(
	ctx context.Context,
	userID string,
	username string,
	password string,
) error {
	var query = `
	INSERT INTO users (user_id, username, password_hash, created_at, updated_at)
	VALUES ($1, $2, $3, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	ON CONFLICT(user_id) DO UPDATE SET
		username=excluded.username,
		password_hash=excluded.password_hash,
		updated_at=CURRENT_TIMESTAMP
	`
	_, err := r.db.ExecContext(ctx, query, userID, username, password)
	return err
}
