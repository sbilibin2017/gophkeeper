package repositories

import (
	"context"
	"testing"
	"time"

	_ "modernc.org/sqlite"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"

	"github.com/sbilibin2017/gophkeeper/internal/models"
)

func setupTestDB2(t *testing.T) *sqlx.DB {
	db, err := sqlx.Open("sqlite", ":memory:")
	require.NoError(t, err)

	schema := `
	CREATE TABLE secrets (
		secret_name TEXT NOT NULL,
		secret_type TEXT NOT NULL,
		secret_owner TEXT NOT NULL,
		ciphertext  BLOB NOT NULL,
		aes_key_enc BLOB NOT NULL,
		created_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY (secret_name, secret_type, secret_owner)
	);
	`
	_, err = db.Exec(schema)
	require.NoError(t, err)

	return db
}

func TestSecretWriteAndRead(t *testing.T) {
	ctx := context.Background()
	db := setupTestDB2(t)
	defer db.Close()

	writeRepo := NewSecretWriteRepository(db)
	readRepo := NewSecretReadRepository(db)

	secret := &models.SecretDB{
		SecretName:  "card1",
		SecretType:  models.SecretTypeBankCard,
		SecretOwner: "alice",
		Ciphertext:  []byte("encrypted-data"),
		AESKeyEnc:   []byte("encrypted-key"),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Save secret
	err := writeRepo.Save(ctx,
		secret.SecretOwner,
		secret.SecretName,
		secret.SecretType,
		secret.Ciphertext,
		secret.AESKeyEnc,
	)
	require.NoError(t, err, "saving secret should not error")

	// Get secret (corrected parameter order: owner, name, type)
	got, err := readRepo.Get(ctx, "alice", "card1", models.SecretTypeBankCard)
	require.NoError(t, err)
	require.Equal(t, secret.SecretName, got.SecretName)
	require.Equal(t, secret.SecretType, got.SecretType)
	require.Equal(t, secret.SecretOwner, got.SecretOwner)
	require.Equal(t, secret.Ciphertext, got.Ciphertext)
	require.Equal(t, secret.AESKeyEnc, got.AESKeyEnc)

	// Update secret ciphertext
	secret.Ciphertext = []byte("updated-encrypted-data")
	err = writeRepo.Save(ctx,
		secret.SecretOwner,
		secret.SecretName,
		secret.SecretType,
		secret.Ciphertext,
		secret.AESKeyEnc,
	)
	require.NoError(t, err)

	// Get updated secret (corrected parameter order)
	gotUpdated, err := readRepo.Get(ctx, "alice", "card1", models.SecretTypeBankCard)
	require.NoError(t, err)
	require.Equal(t, []byte("updated-encrypted-data"), gotUpdated.Ciphertext)

	// List secrets by owner
	list, err := readRepo.List(ctx, "alice")
	require.NoError(t, err)
	require.Len(t, list, 1)
	require.Equal(t, "card1", list[0].SecretName)
}
