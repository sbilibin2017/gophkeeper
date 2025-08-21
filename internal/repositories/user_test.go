package repositories

import (
	"context"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/sbilibin2017/gophkeeper/internal/models"
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
	user := &models.UserDB{
		UserID:       userID,
		Username:     username,
		PasswordHash: passwordHash,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	err := writeRepo.Save(ctx, user)
	assert.NoError(t, err)

	// === Get ===
	userFromDB, err := readRepo.Get(ctx, username)
	assert.NoError(t, err)
	assert.Equal(t, userID, userFromDB.UserID)
	assert.Equal(t, username, userFromDB.Username)
	assert.Equal(t, passwordHash, userFromDB.PasswordHash)

	// === Update ===
	newUsername := "alice2"
	newPasswordHash := "hash456"
	user.Username = newUsername
	user.PasswordHash = newPasswordHash
	user.UpdatedAt = time.Now()

	err = writeRepo.Save(ctx, user)
	assert.NoError(t, err)

	userUpdated, err := readRepo.Get(ctx, newUsername)
	assert.NoError(t, err)
	assert.Equal(t, userID, userUpdated.UserID)
	assert.Equal(t, newUsername, userUpdated.Username)
	assert.Equal(t, newPasswordHash, userUpdated.PasswordHash)
}
