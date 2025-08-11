package repositories

import (
	"context"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite" // SQLite driver
)

// helper to create in-memory DB and create users table
func setupTestDB(t *testing.T) *sqlx.DB {
	db, err := sqlx.Open("sqlite", ":memory:")
	require.NoError(t, err)

	schema := `
	CREATE TABLE users (
		username TEXT PRIMARY KEY,
		password_hash TEXT NOT NULL,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	);
	`
	_, err = db.Exec(schema)
	require.NoError(t, err)

	return db
}

func TestUserWriteRepository_Save_and_UserReadRepository_Get(t *testing.T) {
	ctx := context.Background()
	db := setupTestDB(t)
	defer db.Close()

	writeRepo := NewUserWriteRepository(db)
	readRepo := NewUserReadRepository(db)

	username := "testuser"
	passwordHash := "hash123"

	// Save user
	err := writeRepo.Save(ctx, username, passwordHash)
	require.NoError(t, err)

	// Read user back
	user, err := readRepo.Get(ctx, username)
	require.NoError(t, err)
	assert.Equal(t, username, user.Username)
	assert.Equal(t, passwordHash, user.PasswordHash)
	assert.WithinDuration(t, time.Now(), user.CreatedAt, time.Minute)
	assert.WithinDuration(t, time.Now(), user.UpdatedAt, time.Minute)

	// Sleep 1 second to ensure updated_at timestamp changes on update
	time.Sleep(1 * time.Second)

	// Update user password
	newPasswordHash := "newhash456"
	err = writeRepo.Save(ctx, username, newPasswordHash)
	require.NoError(t, err)

	updatedUser, err := readRepo.Get(ctx, username)
	require.NoError(t, err)
	assert.Equal(t, newPasswordHash, updatedUser.PasswordHash)
	assert.True(t, updatedUser.UpdatedAt.After(user.UpdatedAt))
}

func TestUserReadRepository_Get_NonExistentUser(t *testing.T) {
	ctx := context.Background()
	db := setupTestDB(t)
	defer db.Close()

	readRepo := NewUserReadRepository(db)

	_, err := readRepo.Get(ctx, "nonexistent")
	assert.Error(t, err)
}
