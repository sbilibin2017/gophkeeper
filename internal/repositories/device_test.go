package repositories

import (
	"context"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"

	_ "modernc.org/sqlite"
)

func setupDeviceTestDB(t *testing.T) *sqlx.DB {
	db, err := sqlx.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}

	schema := `
	CREATE TABLE devices (
		device_id TEXT PRIMARY KEY,
		user_id TEXT NOT NULL,
		public_key TEXT NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	`
	_, err = db.Exec(schema)
	if err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}

	return db
}

func TestDeviceWriteAndReadRepositories(t *testing.T) {
	db := setupDeviceTestDB(t)
	defer db.Close()

	ctx := context.Background()
	writeRepo := NewDeviceWriteRepository(db)
	readRepo := NewDeviceReadRepository(db)

	userID := "user1"
	deviceID := "device1"
	publicKey := "pubkey123"

	// === Save ===
	err := writeRepo.Save(ctx, userID, deviceID, publicKey)
	assert.NoError(t, err)

	// === Get ===
	device, err := readRepo.Get(ctx, userID, deviceID)
	assert.NoError(t, err)
	assert.Equal(t, deviceID, device.DeviceID)
	assert.Equal(t, userID, device.UserID)
	assert.Equal(t, publicKey, device.PublicKey)

	// === Update ===
	newPublicKey := "pubkey456"
	err = writeRepo.Save(ctx, userID, deviceID, newPublicKey)
	assert.NoError(t, err)

	deviceUpdated, err := readRepo.Get(ctx, userID, deviceID)
	assert.NoError(t, err)
	assert.Equal(t, newPublicKey, deviceUpdated.PublicKey)
}
