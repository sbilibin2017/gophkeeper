package repositories

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/sbilibin2017/gophkeeper/internal/models"
)

// SecretBinaryClientSaveRepository предоставляет методы для сохранения бинарных секретов клиента в базе данных.
type SecretBinaryClientSaveRepository struct {
	db *sqlx.DB
}

// NewSecretBinaryClientSaveRepository создаёт новый экземпляр SecretBinaryClientSaveRepository.
func NewSecretBinaryClientSaveRepository(db *sqlx.DB) *SecretBinaryClientSaveRepository {
	return &SecretBinaryClientSaveRepository{db: db}
}

// Save сохраняет или обновляет бинарный секрет клиента в базе данных.
//
// Если запись с таким же secret_name уже существует, она будет обновлена.
// Иначе будет вставлена новая запись.
//
// Параметры:
//   - ctx: контекст для управления временем выполнения и отмены операции.
//   - secret: структура с данными бинарного секрета клиента.
//
// Возвращает ошибку, если запрос к базе данных завершился неудачей.
func (r *SecretBinaryClientSaveRepository) Save(
	ctx context.Context,
	secret models.SecretBinaryClient,
) error {
	query := `
		INSERT INTO secret_binary (secret_name, data, meta, updated_at)
		VALUES (:secret_name, :data, :meta, :updated_at)
		ON CONFLICT(secret_name) DO UPDATE SET
			data = excluded.data,
			meta = excluded.meta,
			updated_at = excluded.updated_at;
	`

	_, err := r.db.NamedExecContext(ctx, query, secret)
	return err
}
