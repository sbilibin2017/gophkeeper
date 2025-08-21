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
	device := &models.DeviceDB{
		DeviceID:  deviceID,
		UserID:    userID,
		PublicKey: publicKey,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	err := writeRepo.Save(ctx, device)
	assert.NoError(t, err)

	// === Get ===
	deviceRead, err := readRepo.Get(ctx, userID, deviceID)
	assert.NoError(t, err)
	assert.Equal(t, deviceID, deviceRead.DeviceID)
	assert.Equal(t, userID, deviceRead.UserID)
	assert.Equal(t, publicKey, deviceRead.PublicKey)

	// === Update ===
	newPublicKey := "pubkey456"
	device.PublicKey = newPublicKey
	device.UpdatedAt = time.Now()

	err = writeRepo.Save(ctx, device)
	assert.NoError(t, err)

	deviceUpdated, err := readRepo.Get(ctx, userID, deviceID)
	assert.NoError(t, err)
	assert.Equal(t, newPublicKey, deviceUpdated.PublicKey)
}
