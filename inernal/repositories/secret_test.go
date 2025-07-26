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

func TestSecretWriteRepository_Save(t *testing.T) {
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

	// Проверяем, что данные действительно записались в БД
	var count int
	err = db.GetContext(ctx, &count, `
		SELECT COUNT(*) FROM secrets WHERE secret_name = $1 AND secret_type = $2 AND secret_owner = $3
	`, secret.SecretName, secret.SecretType, secret.SecretOwner)
	require.NoError(t, err)
	require.Equal(t, 1, count)
}

func TestSecretReadRepository_GetAndList(t *testing.T) {
	ctx := context.Background()
	db := setupTestDB(t)
	defer db.Close()

	// Вставим тестовые данные напрямую, чтобы проверить чтение
	secret := &models.SecretDB{
		SecretName:  "test-secret",
		SecretType:  "login",
		SecretOwner: "user123",
		Ciphertext:  []byte("encrypted-data"),
		AESKeyEnc:   []byte("encrypted-key"),
	}
	_, err := db.ExecContext(ctx, `
		INSERT INTO secrets (secret_name, secret_type, secret_owner, ciphertext, aes_key_enc, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`, secret.SecretName, secret.SecretType, secret.SecretOwner, secret.Ciphertext, secret.AESKeyEnc)
	require.NoError(t, err)

	readRepo := NewSecretReadRepository(db)

	// Тест Get
	filterGet := &models.SecretGetFilterDB{
		SecretName:  secret.SecretName,
		SecretType:  secret.SecretType,
		SecretOwner: secret.SecretOwner,
	}
	got, err := readRepo.Get(ctx, filterGet)
	require.NoError(t, err)
	require.Equal(t, secret.SecretName, got.SecretName)
	require.Equal(t, secret.SecretType, got.SecretType)
	require.Equal(t, secret.SecretOwner, got.SecretOwner)
	require.Equal(t, secret.Ciphertext, got.Ciphertext)
	require.Equal(t, secret.AESKeyEnc, got.AESKeyEnc)

	// Тест List
	filterList := &models.SecretListFilterDBRequest{
		SecretOwner: secret.SecretOwner,
	}
	secrets, err := readRepo.List(ctx, filterList)
	require.NoError(t, err)
	require.Len(t, secrets, 1)
	require.Equal(t, "test-secret", secrets[0].SecretName)
}
