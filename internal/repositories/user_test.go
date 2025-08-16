package repositories

import (
	"context"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"

	_ "modernc.org/sqlite"
)

func setupUserTestDB(t *testing.T) *sqlx.DB {
	db, err := sqlx.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}

	schema := `
	CREATE TABLE users (
		user_id TEXT PRIMARY KEY,
		username TEXT NOT NULL UNIQUE,
		password_hash TEXT NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	`
	_, err = db.Exec(schema)
	if err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}

	return db
}

func TestUserWriteAndReadRepositories(t *testing.T) {
	db := setupUserTestDB(t)
	defer db.Close()

	ctx := context.Background()
	writeRepo := NewUserWriteRepository(db)
	readRepo := NewUserReadRepository(db)

	userID := "u1"
	username := "alice"
	passwordHash := "hash123"

	// === Save ===
	err := writeRepo.Save(ctx, userID, username, passwordHash)
	assert.NoError(t, err)

	// === Get ===
	user, err := readRepo.Get(ctx, username)
	assert.NoError(t, err)
	assert.Equal(t, userID, user.UserID)
	assert.Equal(t, username, user.Username)
	assert.Equal(t, passwordHash, user.PasswordHash)

	// === Update ===
	newUsername := "alice2"
	newPasswordHash := "hash456"
	err = writeRepo.Save(ctx, userID, newUsername, newPasswordHash)
	assert.NoError(t, err)

	userUpdated, err := readRepo.Get(ctx, newUsername)
	assert.NoError(t, err)
	assert.Equal(t, userID, userUpdated.UserID)
	assert.Equal(t, newUsername, userUpdated.Username)
	assert.Equal(t, newPasswordHash, userUpdated.PasswordHash)
}
