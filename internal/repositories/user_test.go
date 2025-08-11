package repositories

import (
	"context"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"

	"github.com/sbilibin2017/gophkeeper/internal/models"
	_ "modernc.org/sqlite" // SQLite driver
)

func setupTestDB(t *testing.T) *sqlx.DB {
	t.Helper()

	db, err := sqlx.Open("sqlite", ":memory:")
	assert.NoError(t, err)

	schema := `
	CREATE TABLE users (
		username TEXT PRIMARY KEY,
		password_hash TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	_, err = db.Exec(schema)
	assert.NoError(t, err)

	return db
}

func TestUserWriteRepository_Save(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserWriteRepository(db)
	ctx := context.Background()

	user := &models.UserDB{
		Username:     "testuser",
		PasswordHash: "hash1",
	}

	// Insert new user
	err := repo.Save(ctx, user)
	assert.NoError(t, err)

	// Verify inserted
	var count int
	err = db.GetContext(ctx, &count, "SELECT COUNT(*) FROM users WHERE username = ?", user.Username)
	assert.NoError(t, err)
	assert.Equal(t, 1, count)

	// Update existing user
	user.PasswordHash = "hash2"
	err = repo.Save(ctx, user)
	assert.NoError(t, err)

	// Verify updated
	var updatedHash string
	err = db.GetContext(ctx, &updatedHash, "SELECT password_hash FROM users WHERE username = ?", user.Username)
	assert.NoError(t, err)
	assert.Equal(t, "hash2", updatedHash)
}

func TestUserWriteRepository_Save_Error(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserWriteRepository(db)
	ctx := context.Background()

	// Close DB to force error on query execution
	db.Close()

	user := &models.UserDB{
		Username:     "erroruser",
		PasswordHash: "hash-error",
	}

	err := repo.Save(ctx, user)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to save user")
}

func TestUserReadRepository_Get(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	// Insert test user manually
	_, err := db.ExecContext(ctx, `
		INSERT INTO users (username, password_hash)
		VALUES (?, ?)`, "readuser", "hash123")
	assert.NoError(t, err)

	repo := NewUserReadRepository(db)

	// Get existing user
	user, err := repo.Get(ctx, "readuser")
	assert.NoError(t, err)
	assert.Equal(t, "readuser", user.Username)
	assert.Equal(t, "hash123", user.PasswordHash)

	// Try to get non-existing user
	user, err = repo.Get(ctx, "nouser")
	assert.Error(t, err)
	assert.Nil(t, user)
}

func TestUserReadRepository_Get_Error(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserReadRepository(db)
	ctx := context.Background()

	// Close DB to force error on query
	db.Close()

	user, err := repo.Get(ctx, "anyuser")
	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Contains(t, err.Error(), "failed to get user")
}
