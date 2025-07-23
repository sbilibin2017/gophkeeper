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

func setupTestDB(t *testing.T) *sqlx.DB {
	db, err := sqlx.Open("sqlite", ":memory:")
	require.NoError(t, err)

	schema := `
	CREATE TABLE binaries (
		secret_name TEXT PRIMARY KEY,
		secret_owner TEXT NOT NULL,
		file_path TEXT NOT NULL,
		data BLOB NOT NULL,
		meta TEXT,
		updated_at DATETIME NOT NULL
	);
	`
	_, err = db.Exec(schema)
	require.NoError(t, err)

	return db
}

func TestBinaryWriteAndReadRepositories(t *testing.T) {
	db := setupTestDB(t)
	writeRepo := NewBinaryWriteRepository(db)
	readRepo := NewBinaryReadRepository(db)

	ctx := context.Background()
	meta := "meta info"
	now := time.Now()

	binarySecret := &models.Binary{
		SecretName:  "bin1",
		SecretOwner: "user1",
		FilePath:    "/tmp/file1",
		Data:        []byte{1, 2, 3, 4},
		Meta:        &meta,
		UpdatedAt:   now,
	}

	// Add binary secret
	err := writeRepo.Add(ctx, binarySecret)
	require.NoError(t, err)

	// Read and verify
	binaries, err := readRepo.List(ctx)
	require.NoError(t, err)
	require.Len(t, binaries, 1)

	got := binaries[0]
	require.Equal(t, binarySecret.SecretName, got.SecretName)
	require.Equal(t, binarySecret.SecretOwner, got.SecretOwner)
	require.Equal(t, binarySecret.FilePath, got.FilePath)
	require.Equal(t, binarySecret.Data, got.Data)
	require.Equal(t, binarySecret.Meta, got.Meta)
	require.WithinDuration(t, binarySecret.UpdatedAt, got.UpdatedAt, time.Second)
}

func TestBinaryListEmpty(t *testing.T) {
	db := setupTestDB(t)
	readRepo := NewBinaryReadRepository(db)

	ctx := context.Background()
	binaries, err := readRepo.List(ctx)
	require.NoError(t, err)
	require.Empty(t, binaries)
}
