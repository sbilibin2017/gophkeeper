package repositories

import (
	"context"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"

	_ "modernc.org/sqlite"
)

func setupSecretKeyTestDB(t *testing.T) *sqlx.DB {
	db, err := sqlx.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}

	schema := `
	CREATE TABLE secret_keys (
		secret_key_id TEXT PRIMARY KEY,
		secret_id TEXT NOT NULL,
		device_id TEXT NOT NULL,
		encrypted_aes_key BLOB NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(secret_id, device_id)
	);
	`
	_, err = db.Exec(schema)
	if err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}

	return db
}

func TestSecretKeyWriteAndReadRepositories(t *testing.T) {
	db := setupSecretKeyTestDB(t)
	defer db.Close()

	ctx := context.Background()
	writeRepo := NewSecretKeyWriteRepository(db)
	readRepo := NewSecretKeyReadRepository(db)

	secretKeyID := "key1"
	secretID := "secret1"
	deviceID := "device1"
	encryptedAESKey := []byte("aeskey123")

	// === Save ===
	err := writeRepo.Save(ctx, secretKeyID, secretID, deviceID, encryptedAESKey)
	assert.NoError(t, err)

	// === Get ===
	key, err := readRepo.Get(ctx, secretID, deviceID)
	assert.NoError(t, err)
	assert.Equal(t, secretKeyID, key.SecretKeyID)
	assert.Equal(t, secretID, key.SecretID)
	assert.Equal(t, deviceID, key.DeviceID)
	assert.Equal(t, encryptedAESKey, key.EncryptedAESKey)

	// === Update ===
	newAESKey := []byte("newaeskey")
	err = writeRepo.Save(ctx, secretKeyID, secretID, deviceID, newAESKey)
	assert.NoError(t, err)

	keyUpdated, err := readRepo.Get(ctx, secretID, deviceID)
	assert.NoError(t, err)
	assert.Equal(t, newAESKey, keyUpdated.EncryptedAESKey)
}
