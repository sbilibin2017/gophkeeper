package repositories

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/sbilibin2017/gophkeeper/internal/models"
)

// DeviceReaderRepository предоставляет методы чтения данных устройства из БД.
type DeviceReaderRepository struct {
	db *sqlx.DB
}

// NewDeviceReaderRepository создаёт новый экземпляр DeviceReaderRepository.
func NewDeviceReaderRepository(db *sqlx.DB) *DeviceReaderRepository {
	return &DeviceReaderRepository{db: db}
}

// GetByID возвращает устройство по deviceID.
func (r *DeviceReaderRepository) GetByID(ctx context.Context, deviceID string) (*models.DeviceDB, error) {
	var query = "SELECT * FROM devices WHERE device_id=$1"
	var device models.DeviceDB
	err := r.db.GetContext(ctx, &device, query, deviceID)
	if err != nil {
		return nil, err
	}
	return &device, nil
}

// DeviceWriterRepository предоставляет методы записи данных устройства в БД.
type DeviceWriterRepository struct {
	db *sqlx.DB
}

// NewDeviceWriterRepository создаёт новый экземпляр DeviceWriterRepository.
func NewDeviceWriterRepository(db *sqlx.DB) *DeviceWriterRepository {
	return &DeviceWriterRepository{db: db}
}

// Save создаёт новое устройство или обновляет существующее.
func (r *DeviceWriterRepository) Save(
	ctx context.Context,
	deviceID string,
	userID string,
	publicKey string,
	encryptedDEK string,
) error {
	var query = `
	INSERT INTO devices (device_id, user_id, public_key, encrypted_dek, created_at, updated_at)
	VALUES ($1, $2, $3, $4, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	ON CONFLICT(device_id) DO UPDATE SET		
		public_key=excluded.public_key,
		encrypted_dek=excluded.encrypted_dek,
		updated_at=CURRENT_TIMESTAMP
	`
	_, err := r.db.ExecContext(ctx, query, deviceID, userID, publicKey, encryptedDEK)
	return err
}
