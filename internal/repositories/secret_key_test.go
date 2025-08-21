package repositories

import (
	"context"
	"encoding/base64"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/sbilibin2017/gophkeeper/internal/models"
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
		secret_id TEXT NOT NULL,
		device_id TEXT NOT NULL,
		encrypted_aes_key TEXT NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY (secret_id, device_id)
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

	secretID := "secret1"
	deviceID := "device1"
	encryptedAESKey := []byte("aeskey123")

	// === Save ===
	key := &models.SecretKeyDB{
		SecretID:        secretID,
		DeviceID:        deviceID,
		EncryptedAESKey: base64.StdEncoding.EncodeToString(encryptedAESKey),
	}
	err := writeRepo.Save(ctx, key)
	assert.NoError(t, err)

	// === Get ===
	keyRead, err := readRepo.Get(ctx, secretID, deviceID)
	assert.NoError(t, err)

	// Декодируем обратно из Base64
	aesKeyDecoded, err := base64.StdEncoding.DecodeString(keyRead.EncryptedAESKey)
	assert.NoError(t, err)
	assert.Equal(t, encryptedAESKey, aesKeyDecoded)

	// === Update ===
	newAESKey := []byte("newaeskey")
	key.EncryptedAESKey = base64.StdEncoding.EncodeToString(newAESKey)
	err = writeRepo.Save(ctx, key)
	assert.NoError(t, err)

	keyUpdated, err := readRepo.Get(ctx, secretID, deviceID)
	assert.NoError(t, err)
	newAESKeyDecoded, err := base64.StdEncoding.DecodeString(keyUpdated.EncryptedAESKey)
	assert.NoError(t, err)
	assert.Equal(t, newAESKey, newAESKeyDecoded)
}
