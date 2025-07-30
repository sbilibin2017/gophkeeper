package repositories

import (
	"context"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"

	"github.com/sbilibin2017/gophkeeper/internal/models"
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
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY (secret_name, secret_type, secret_owner)
	);
	`
	_, err = db.Exec(schema)
	require.NoError(t, err)

	return db
}

func TestSecretWriteRepository_SaveAndGet(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	writeRepo := NewSecretWriteRepository(db)
	readRepo := NewSecretReadRepository(db)

	ctx := context.Background()
	owner := "user1"

	secretName := "secret1"
	secretType := models.SecretTypeText
	ciphertext := []byte("SecretEncrypted-data")
	aesKeyEnc := []byte("SecretEncrypted-key")

	// Save new secret
	err := writeRepo.Save(ctx, owner, secretName, secretType, ciphertext, aesKeyEnc)
	require.NoError(t, err)

	// Get secret and verify
	got, err := readRepo.Get(ctx, owner, secretType, secretName)
	require.NoError(t, err)
	assert.Equal(t, secretName, got.SecretName)
	assert.Equal(t, secretType, got.SecretType)
	assert.Equal(t, owner, got.SecretOwner)
	assert.Equal(t, ciphertext, got.Ciphertext)
	assert.Equal(t, aesKeyEnc, got.AESKeyEnc)

	// Wait to ensure updated_at changes (SQLite timestamps have seconds precision)
	timeBeforeUpdate := time.Now()
	time.Sleep(1 * time.Second) // sleep 1 second to let the timestamp advance

	// Update secret
	updatedCiphertext := []byte("updated-SecretEncrypted-data")
	err = writeRepo.Save(ctx, owner, secretName, secretType, updatedCiphertext, aesKeyEnc)
	require.NoError(t, err)

	gotUpdated, err := readRepo.Get(ctx, owner, secretType, secretName)
	require.NoError(t, err)
	assert.Equal(t, updatedCiphertext, gotUpdated.Ciphertext)
	// updated_at should be after timeBeforeUpdate
	assert.True(t, gotUpdated.UpdatedAt.After(timeBeforeUpdate) || gotUpdated.UpdatedAt.Equal(timeBeforeUpdate))
}

func TestSecretReadRepository_List(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	writeRepo := NewSecretWriteRepository(db)
	readRepo := NewSecretReadRepository(db)

	ctx := context.Background()
	owner := "user1"

	secrets := []struct {
		SecretName string
		SecretType string
		Ciphertext []byte
		AESKeyEnc  []byte
	}{
		{"secret1", models.SecretTypeText, []byte("data1"), []byte("key1")},
		{"secret2", models.SecretTypeBankCard, []byte("data2"), []byte("key2")},
	}

	for _, s := range secrets {
		err := writeRepo.Save(ctx, owner, s.SecretName, s.SecretType, s.Ciphertext, s.AESKeyEnc)
		require.NoError(t, err)
	}

	gotSecrets, err := readRepo.List(ctx, owner)
	require.NoError(t, err)
	assert.Len(t, gotSecrets, len(secrets))

	for _, expected := range secrets {
		found := false
		for _, got := range gotSecrets {
			if got.SecretName == expected.SecretName && got.SecretType == expected.SecretType {
				assert.Equal(t, expected.Ciphertext, got.Ciphertext)
				assert.Equal(t, expected.AESKeyEnc, got.AESKeyEnc)
				found = true
				break
			}
		}
		assert.True(t, found, "secret not found: %s", expected.SecretName)
	}
}
