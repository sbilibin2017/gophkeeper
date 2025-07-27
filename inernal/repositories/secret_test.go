package repositories

import (
	"context"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"

	"github.com/sbilibin2017/gophkeeper/inernal/models"
)

func setupTestDB(t *testing.T) *sqlx.DB {
	db, err := sqlx.Open("sqlite", ":memory:")
	require.NoError(t, err)

	schema := `
	CREATE TABLE secrets (
		secret_name TEXT NOT NULL,
		secret_type TEXT NOT NULL,
		secret_owner TEXT NOT NULL,
		ciphertext BLOB NOT NULL,
		aes_key_enc BLOB NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY(secret_name, secret_type, secret_owner)
	);`

	_, err = db.Exec(schema)
	require.NoError(t, err)

	return db
}

func TestSaveSecret(t *testing.T) {
	ctx := context.Background()
	db := setupTestDB(t)
	defer db.Close()

	writeRepo := NewSecretWriteRepository(db)

	secret := &models.SecretDB{
		SecretName:  "test-secret",
		SecretType:  "login",
		SecretOwner: "user123",
		Ciphertext:  []byte("encrypted-data"),
		AESKeyEnc:   []byte("encrypted-key"),
	}

	err := writeRepo.Save(ctx, secret)
	require.NoError(t, err)

	// Verify the secret was stored
	var count int
	err = db.GetContext(ctx, &count, `
		SELECT COUNT(*) FROM secrets WHERE secret_name = ? AND secret_type = ? AND secret_owner = ?
	`, secret.SecretName, secret.SecretType, secret.SecretOwner)
	require.NoError(t, err)
	require.Equal(t, 1, count)
}

func TestGetSecretAndListSecrets(t *testing.T) {
	ctx := context.Background()
	db := setupTestDB(t)
	defer db.Close()

	writeRepo := NewSecretWriteRepository(db)
	readRepo := NewSecretReadRepository(db)

	secret := &models.SecretDB{
		SecretName:  "test-secret",
		SecretType:  "login",
		SecretOwner: "user123",
		Ciphertext:  []byte("encrypted-data"),
		AESKeyEnc:   []byte("encrypted-key"),
	}

	// Use repository to insert
	err := writeRepo.Save(ctx, secret)
	require.NoError(t, err)

	// Test Get
	got, err := readRepo.Get(ctx, secret.SecretName, secret.SecretType, secret.SecretOwner)
	require.NoError(t, err)
	require.Equal(t, secret.SecretName, got.SecretName)
	require.Equal(t, secret.SecretType, got.SecretType)
	require.Equal(t, secret.SecretOwner, got.SecretOwner)
	require.Equal(t, secret.Ciphertext, got.Ciphertext)
	require.Equal(t, secret.AESKeyEnc, got.AESKeyEnc)

	// Test List
	secrets, err := readRepo.List(ctx, secret.SecretOwner)
	require.NoError(t, err)
	require.Len(t, secrets, 1)
	require.Equal(t, secret.SecretName, secrets[0].SecretName)
	require.Equal(t, secret.SecretType, secrets[0].SecretType)
}
