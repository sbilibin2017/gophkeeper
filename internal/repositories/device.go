package repositories

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/sbilibin2017/gophkeeper/internal/models"
)

// DeviceWriteRepository управляет записью устройств пользователей
type DeviceWriteRepository struct {
	db *sqlx.DB
}

// NewDeviceWriteRepository создаёт новый репозиторий записи устройств
func NewDeviceWriteRepository(db *sqlx.DB) *DeviceWriteRepository {
	return &DeviceWriteRepository{db: db}
}

// Save вставляет новое устройство или обновляет существующее по device_id
func (r *DeviceWriteRepository) Save(
	ctx context.Context,
	device *models.DeviceDB,
) error {
	query := `
	INSERT INTO devices (device_id, user_id, public_key, created_at, updated_at)
	VALUES ($1, $2, $3, $4, $5)
	ON CONFLICT(device_id) DO UPDATE SET
		user_id = EXCLUDED.user_id,
		public_key = EXCLUDED.public_key,
		updated_at = EXCLUDED.updated_at
	`
	_, err := r.db.ExecContext(
		ctx,
		query,
		device.DeviceID,
		device.UserID,
		device.PublicKey,
		device.CreatedAt,
		device.UpdatedAt,
	)
	return err
}

// DeviceReadRepository управляет чтением устройств пользователей
type DeviceReadRepository struct {
	db *sqlx.DB
}

// NewDeviceReadRepository создаёт новый репозиторий чтения устройств
func NewDeviceReadRepository(db *sqlx.DB) *DeviceReadRepository {
	return &DeviceReadRepository{db: db}
}

// Get возвращает устройство по userID и deviceID
func (r *DeviceReadRepository) Get(
	ctx context.Context,
	userID, deviceID string,
) (*models.DeviceDB, error) {
	var device models.DeviceDB
	query := `SELECT device_id, user_id, public_key, created_at, updated_at FROM devices WHERE user_id = $1 AND device_id = $2`
	if err := r.db.GetContext(ctx, &device, query, userID, deviceID); err != nil {
		return nil, err
	}
	return &device, nil
}
