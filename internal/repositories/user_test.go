package repositories

import (
	"context"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/sbilibin2017/gophkeeper/internal/models"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func setupTestDBForUsers(t *testing.T) *sqlx.DB {
	db, err := sqlx.Open("sqlite", ":memory:")
	require.NoError(t, err)

	schema := `
	CREATE TABLE users (
		secret_name TEXT PRIMARY KEY,
		secret_owner TEXT NOT NULL,
		login TEXT NOT NULL,
		password TEXT NOT NULL,
		meta TEXT,
		updated_at DATETIME NOT NULL
	);
	`
	_, err = db.Exec(schema)
	require.NoError(t, err)

	return db
}

func TestUserWriteAndReadRepositories(t *testing.T) {
	db := setupTestDBForUsers(t)
	writeRepo := NewUserWriteRepository(db)
	readRepo := NewUserReadRepository(db)

	ctx := context.Background()
	meta := "user meta info"
	now := time.Now()

	userSecret := &models.User{
		SecretName:  "user1",
		SecretOwner: "owner1",
		Login:       "login1",
		Password:    "password1",
		Meta:        &meta,
		UpdatedAt:   now,
	}

	// Add user secret
	err := writeRepo.Add(ctx, userSecret)
	require.NoError(t, err)

	// Read and verify
	users, err := readRepo.List(ctx)
	require.NoError(t, err)
	require.Len(t, users, 1)

	got := users[0]
	require.Equal(t, userSecret.SecretName, got.SecretName)
	require.Equal(t, userSecret.SecretOwner, got.SecretOwner)
	require.Equal(t, userSecret.Login, got.Login)
	require.Equal(t, userSecret.Password, got.Password)
	require.Equal(t, userSecret.Meta, got.Meta)
	require.WithinDuration(t, userSecret.UpdatedAt, got.UpdatedAt, time.Second)
}

func TestUserListEmpty(t *testing.T) {
	db := setupTestDBForUsers(t)
	readRepo := NewUserReadRepository(db)

	ctx := context.Background()
	users, err := readRepo.List(ctx)
	require.NoError(t, err)
	require.Empty(t, users)
}
