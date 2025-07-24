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

func openTestDB(t *testing.T) *sqlx.DB {
	db, err := sqlx.Open("sqlite", ":memory:")
	require.NoError(t, err)

	err = CreateEncryptedSecretsTable(context.Background(), db)
	require.NoError(t, err)

	return db
}

func TestEncryptedSecretWriteRepository_SaveAndDelete(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	writeRepo := NewEncryptedSecretWriteRepository(db)
	readRepo := NewEncryptedSecretReadRepository(db)

	ctx := context.Background()

	secret := &models.EncryptedSecret{
		SecretName: "testSecret",
		SecretType: "password",
		Ciphertext: []byte("encrypteddata"),
		AESKeyEnc:  []byte("encryptedkey"),
		Timestamp:  time.Now().Unix(),
	}

	// Save secret
	err := writeRepo.Save(ctx, secret)
	require.NoError(t, err)

	// Retrieve secret and verify fields
	got, err := readRepo.Get(ctx, "testSecret")
	require.NoError(t, err)

	assert.Equal(t, secret.SecretName, got.SecretName)
	assert.Equal(t, secret.SecretType, got.SecretType)
	assert.Equal(t, secret.Ciphertext, got.Ciphertext)
	assert.Equal(t, secret.AESKeyEnc, got.AESKeyEnc)
	assert.Equal(t, secret.Timestamp, got.Timestamp)

	// Delete secret
	err = writeRepo.Delete(ctx, "testSecret")
	require.NoError(t, err)

	// Try to get deleted secret — should error
	_, err = readRepo.Get(ctx, "testSecret")
	assert.Error(t, err)
}

func TestEncryptedSecretReadRepository_List(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	writeRepo := NewEncryptedSecretWriteRepository(db)
	readRepo := NewEncryptedSecretReadRepository(db)

	ctx := context.Background()

	secrets := []*models.EncryptedSecret{
		{
			SecretName: "secret1",
			SecretType: "type1",
			Ciphertext: []byte("data1"),
			AESKeyEnc:  []byte("key1"),
			Timestamp:  time.Now().Unix(),
		},
		{
			SecretName: "secret2",
			SecretType: "type2",
			Ciphertext: []byte("data2"),
			AESKeyEnc:  []byte("key2"),
			Timestamp:  time.Now().Unix(),
		},
	}

	for _, s := range secrets {
		require.NoError(t, writeRepo.Save(ctx, s))
	}

	gotSecrets, err := readRepo.List(ctx)
	require.NoError(t, err)

	assert.Len(t, gotSecrets, len(secrets))

	// Map by name for easier assertions
	gotMap := make(map[string]*models.EncryptedSecret)
	for _, gs := range gotSecrets {
		gotMap[gs.SecretName] = gs
	}

	for _, expected := range secrets {
		got, ok := gotMap[expected.SecretName]
		assert.True(t, ok)
		assert.Equal(t, expected.SecretType, got.SecretType)
		assert.Equal(t, expected.Ciphertext, got.Ciphertext)
		assert.Equal(t, expected.AESKeyEnc, got.AESKeyEnc)
		assert.Equal(t, expected.Timestamp, got.Timestamp)
	}
}

func TestDropEncryptedSecretsTable(t *testing.T) {
	ctx := context.Background()

	// Connect to in-memory SQLite database
	db, err := sqlx.ConnectContext(ctx, "sqlite", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	// Create table first to ensure Drop works on existing table
	err = CreateEncryptedSecretsTable(ctx, db)
	require.NoError(t, err)

	// Drop the table
	err = DropEncryptedSecretsTable(ctx, db)
	require.NoError(t, err)

	// Try dropping again to provoke error (optional, just to check)
	err = DropEncryptedSecretsTable(ctx, db)
	require.Error(t, err)
}
