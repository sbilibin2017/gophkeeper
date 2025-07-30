package repositories

import (
	"context"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func setupUserTestDB(t *testing.T) *sqlx.DB {
	db, err := sqlx.Open("sqlite", ":memory:")
	require.NoError(t, err)

	schema := `
	CREATE TABLE users (
		username TEXT PRIMARY KEY,
		password_hash TEXT NOT NULL,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);
	`
	_, err = db.Exec(schema)
	require.NoError(t, err)

	return db
}

func TestUserWriteRepository_SaveAndGet(t *testing.T) {
	db := setupUserTestDB(t)
	defer db.Close()

	writeRepo := NewUserWriteRepository(db)
	readRepo := NewUserReadRepository(db)

	ctx := context.Background()
	username := "testuser"
	passwordHash := "hash123"

	// Save new user
	err := writeRepo.Save(ctx, username, passwordHash)
	require.NoError(t, err)

	// Get user
	got, err := readRepo.Get(ctx, username)
	require.NoError(t, err)
	assert.Equal(t, username, got.Username)
	assert.Equal(t, passwordHash, got.PasswordHash)

	// Capture time before update
	timeBeforeUpdate := time.Now()
	time.Sleep(1 * time.Second)

	// Update password hash
	newPasswordHash := "updated-hash456"
	err = writeRepo.Save(ctx, username, newPasswordHash)
	require.NoError(t, err)

	// Get updated user
	updated, err := readRepo.Get(ctx, username)
	require.NoError(t, err)
	assert.Equal(t, newPasswordHash, updated.PasswordHash)
	assert.True(t, updated.UpdatedAt.After(timeBeforeUpdate) || updated.UpdatedAt.Equal(timeBeforeUpdate))
}
