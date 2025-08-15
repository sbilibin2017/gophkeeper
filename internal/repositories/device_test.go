package repositories

import (
	"context"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	_ "modernc.org/sqlite"
)

func setupDeviceTestDB(t *testing.T) *sqlx.DB {
	db, err := sqlx.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}

	schema := `
	CREATE TABLE devices (
		device_id TEXT PRIMARY KEY,
		user_id TEXT NOT NULL,
		public_key TEXT NOT NULL,
		encrypted_dek TEXT NOT NULL,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL
	);
	`
	_, err = db.Exec(schema)
	if err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}

	return db
}

func TestDeviceWriterRepository_Save(t *testing.T) {
	db := setupDeviceTestDB(t)
	defer db.Close()

	writer := NewDeviceWriterRepository(db)
	ctx := context.Background()

	// Вставка нового устройства
	err := writer.Save(ctx, "device1", "user1", "pubkey123", "dek123")
	assert.NoError(t, err)

	var count int
	err = db.Get(&count, "SELECT COUNT(*) FROM devices WHERE device_id=?", "device1")
	assert.NoError(t, err)
	assert.Equal(t, 1, count)

	// Проверка обновления существующего устройства
	err = writer.Save(ctx, "device1", "user1", "pubkey456", "dek456")
	assert.NoError(t, err)

	var pubkey, dek string
	err = db.Get(&pubkey, "SELECT public_key FROM devices WHERE device_id=?", "device1")
	assert.NoError(t, err)
	assert.Equal(t, "pubkey456", pubkey)

	err = db.Get(&dek, "SELECT encrypted_dek FROM devices WHERE device_id=?", "device1")
	assert.NoError(t, err)
	assert.Equal(t, "dek456", dek)
}

func TestDeviceReaderRepository_GetByID(t *testing.T) {
	db := setupDeviceTestDB(t)
	defer db.Close()

	// Вставляем тестовое устройство
	_, err := db.Exec(
		"INSERT INTO devices (device_id, user_id, public_key, encrypted_dek, created_at, updated_at) VALUES (?, ?, ?, ?, datetime('now'), datetime('now'))",
		"device1", "user1", "pubkey123", "dek123",
	)
	assert.NoError(t, err)

	reader := NewDeviceReaderRepository(db)
	ctx := context.Background()

	// Проверяем существующее устройство
	device, err := reader.GetByID(ctx, "device1")
	assert.NoError(t, err)
	assert.NotNil(t, device)
	assert.Equal(t, "device1", device.DeviceID)
	assert.Equal(t, "user1", device.UserID)
	assert.Equal(t, "pubkey123", device.PublicKey)
	assert.Equal(t, "dek123", device.EncryptedDEK)
	assert.WithinDuration(t, time.Now(), device.CreatedAt, time.Second)
	assert.WithinDuration(t, time.Now(), device.UpdatedAt, time.Second)

	// Проверяем отсутствие устройства
	device, err = reader.GetByID(ctx, "device2")
	assert.Error(t, err)
	assert.Nil(t, device)
}
