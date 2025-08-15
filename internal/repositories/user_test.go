package repositories

import (
	"context"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	_ "modernc.org/sqlite"
)

func setupTestDB(t *testing.T) *sqlx.DB {
	db, err := sqlx.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}

	schema := `
	CREATE TABLE users (
		user_id TEXT PRIMARY KEY,
		username TEXT NOT NULL,
		password_hash TEXT NOT NULL,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL
	);
	`
	_, err = db.Exec(schema)
	if err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}

	return db
}

func TestUserWriterRepository_Save(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	writer := NewUserWriterRepository(db)
	ctx := context.Background()

	err := writer.Save(ctx, "user1", "alice", "password123")
	assert.NoError(t, err)

	// Проверяем вставку
	var count int
	err = db.Get(&count, "SELECT COUNT(*) FROM users WHERE user_id=?", "user1")
	assert.NoError(t, err)
	assert.Equal(t, 1, count)

	// Проверяем обновление
	err = writer.Save(ctx, "user1", "alice_updated", "password456")
	assert.NoError(t, err)

	var username, password string
	err = db.Get(&username, "SELECT username FROM users WHERE user_id=?", "user1")
	assert.NoError(t, err)
	assert.Equal(t, "alice_updated", username)

	err = db.Get(&password, "SELECT password_hash FROM users WHERE user_id=?", "user1")
	assert.NoError(t, err)
	assert.Equal(t, "password456", password)
}

func TestUserReaderRepository_GetByUsername(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Вставляем тестового пользователя
	_, err := db.Exec(
		"INSERT INTO users (user_id, username, password_hash, created_at, updated_at) VALUES (?, ?, ?, datetime('now'), datetime('now'))",
		"user1", "alice", "password123",
	)
	assert.NoError(t, err)

	reader := NewUserReaderRepository(db)
	ctx := context.Background()

	user, err := reader.GetByUsername(ctx, "alice")
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, "alice", user.Username)
	assert.Equal(t, "password123", user.PasswordHash)

	// Проверяем случай отсутствия пользователя
	user, err = reader.GetByUsername(ctx, "bob")
	assert.Error(t, err)
	assert.Nil(t, user)
}
