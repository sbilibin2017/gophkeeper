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

// SecretBinaryClientListRepository предоставляет методы для чтения бинарных секретов клиента из базы данных.
type SecretBinaryClientListRepository struct {
	db *sqlx.DB
}

// NewSecretBinaryClientListRepository создаёт новый экземпляр SecretBinaryClientListRepository.
func NewSecretBinaryClientListRepository(db *sqlx.DB) *SecretBinaryClientListRepository {
	return &SecretBinaryClientListRepository{db: db}
}

// List возвращает все бинарные секреты клиента из базы данных.
//
// Параметры:
//   - ctx: контекст для управления временем выполнения и отмены операции.
//
// Возвращает:
//   - []models.SecretBinaryClient: список всех найденных бинарных секретов
//   - error: ошибка, если операция завершилась неудачей
func (r *SecretBinaryClientListRepository) List(ctx context.Context) ([]models.SecretBinaryClient, error) {
	query := `
		SELECT secret_name, data, meta, updated_at
		FROM secret_binary_client;
	`

	var secrets []models.SecretBinaryClient
	if err := r.db.SelectContext(ctx, &secrets, query); err != nil {
		return nil, err
	}

	return secrets, nil
}
