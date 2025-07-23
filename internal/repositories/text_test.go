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

func setupTestDBForTexts(t *testing.T) *sqlx.DB {
	db, err := sqlx.Open("sqlite", ":memory:")
	require.NoError(t, err)

	schema := `
	CREATE TABLE texts (
		secret_name TEXT PRIMARY KEY,
		secret_owner TEXT NOT NULL,
		data TEXT NOT NULL,
		meta TEXT,
		updated_at DATETIME NOT NULL
	);
	`
	_, err = db.Exec(schema)
	require.NoError(t, err)

	return db
}

func TestTextWriteAndReadRepositories(t *testing.T) {
	db := setupTestDBForTexts(t)
	writeRepo := NewTextWriteRepository(db)
	readRepo := NewTextReadRepository(db)

	ctx := context.Background()
	meta := "some meta"
	now := time.Now()

	textSecret := &models.Text{
		SecretName:  "text1",
		SecretOwner: "user1",
		Data:        "This is a secret text",
		Meta:        &meta,
		UpdatedAt:   now,
	}

	// Add text secret
	err := writeRepo.Add(ctx, textSecret)
	require.NoError(t, err)

	// Read and verify
	texts, err := readRepo.List(ctx)
	require.NoError(t, err)
	require.Len(t, texts, 1)

	got := texts[0]
	require.Equal(t, textSecret.SecretName, got.SecretName)
	require.Equal(t, textSecret.SecretOwner, got.SecretOwner)
	require.Equal(t, textSecret.Data, got.Data)
	require.Equal(t, textSecret.Meta, got.Meta)
	require.WithinDuration(t, textSecret.UpdatedAt, got.UpdatedAt, time.Second)
}

func TestTextListEmpty(t *testing.T) {
	db := setupTestDBForTexts(t)
	readRepo := NewTextReadRepository(db)

	ctx := context.Background()
	texts, err := readRepo.List(ctx)
	require.NoError(t, err)
	require.Empty(t, texts)
}
