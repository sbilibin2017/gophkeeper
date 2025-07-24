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

func setupTestDB(t *testing.T) *sqlx.DB {
	db, err := sqlx.Open("sqlite", ":memory:")
	assert.NoError(t, err)
	t.Cleanup(func() { db.Close() })

	err = CreateEncryptedSecretsTable(context.Background(), db)
	assert.NoError(t, err)

	return db
}

func TestCreateAndDropEncryptedSecretsTable(t *testing.T) {
	db, err := sqlx.Open("sqlite", ":memory:")
	assert.NoError(t, err)
	defer db.Close()

	ctx := context.Background()

	// Create table
	err = CreateEncryptedSecretsTable(ctx, db)
	assert.NoError(t, err)

	// Drop table
	err = DropEncryptedSecretsTable(ctx, db)
	assert.NoError(t, err)
}

func TestEncryptedSecretWriteRepository_SaveAndDelete(t *testing.T) {
	db := setupTestDB(t)
	repo := NewEncryptedSecretWriteRepository(db)
	ctx := context.Background()

	now := time.Now().UnixNano()
	secret := &models.EncryptedSecret{
		SecretName: "mysecret",
		SecretType: "text",
		Ciphertext: []byte("encrypteddata"),
		AESKeyEnc:  []byte("encryptedkey"),
		Timestamp:  now,
	}

	// Save new secret
	err := repo.Save(ctx, secret)
	assert.NoError(t, err)

	// Save again to update
	secret.Ciphertext = []byte("updatedcipher")
	err = repo.Save(ctx, secret)
	assert.NoError(t, err)

}

func TestEncryptedSecretWriteRepository_Save_Error(t *testing.T) {
	db, err := sqlx.Open("sqlite", ":memory:")
	assert.NoError(t, err)
	defer db.Close()

	repo := NewEncryptedSecretWriteRepository(db)
	ctx := context.Background()

	secret := &models.EncryptedSecret{
		SecretName: "failsecret",
		SecretType: "text",
		Ciphertext: []byte("data"),
		AESKeyEnc:  []byte("key"),
		Timestamp:  time.Now().UnixNano(),
	}

	// Without creating table, Save should fail
	err = repo.Save(ctx, secret)
	assert.Error(t, err)
}

func TestEncryptedSecretReadRepository_GetAndList(t *testing.T) {
	db := setupTestDB(t)
	writeRepo := NewEncryptedSecretWriteRepository(db)
	readRepo := NewEncryptedSecretReadRepository(db)
	ctx := context.Background()

	now := time.Now().UnixNano()

	secrets := []*models.EncryptedSecret{
		{
			SecretName: "secret1",
			SecretType: "text",
			Ciphertext: []byte("data1"),
			AESKeyEnc:  []byte("key1"),
			Timestamp:  now,
		},
		{
			SecretName: "secret2",
			SecretType: "binary",
			Ciphertext: []byte("data2"),
			AESKeyEnc:  []byte("key2"),
			Timestamp:  now,
		},
	}

	for _, s := range secrets {
		err := writeRepo.Save(ctx, s)
		assert.NoError(t, err)
	}

	// Test Get existing secret
	got, err := readRepo.Get(ctx, "secret1")
	assert.NoError(t, err)
	assert.Equal(t, "secret1", got.SecretName)
	assert.Equal(t, "text", got.SecretType)
	assert.Equal(t, []byte("data1"), got.Ciphertext)

	// Test Get non-existing secret
	_, err = readRepo.Get(ctx, "nonexistent")
	assert.Error(t, err)

	// Test List all secrets
	allSecrets, err := readRepo.List(ctx)
	assert.NoError(t, err)
	assert.Len(t, allSecrets, 2)
	names := []string{allSecrets[0].SecretName, allSecrets[1].SecretName}
	assert.Contains(t, names, "secret1")
	assert.Contains(t, names, "secret2")
}

func TestEncryptedSecretReadRepository_List_Error(t *testing.T) {
	db, err := sqlx.Open("sqlite", ":memory:")
	assert.NoError(t, err)
	defer db.Close()

	readRepo := NewEncryptedSecretReadRepository(db)
	ctx := context.Background()

	// No table created, List should fail
	_, err = readRepo.List(ctx)
	assert.Error(t, err)
}
