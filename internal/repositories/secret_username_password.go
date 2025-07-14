package repositories

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/sbilibin2017/gophkeeper/internal/models"
)

// SecretUsernamePasswordClientSaveRepository отвечает за сохранение секретов с логином и паролем клиента в базе данных.
type SecretUsernamePasswordClientSaveRepository struct {
	db *sqlx.DB
}

// NewSecretUsernamePasswordClientSaveRepository создаёт новый репозиторий для работы с секретами типа "логин-пароль".
func NewSecretUsernamePasswordClientSaveRepository(db *sqlx.DB) *SecretUsernamePasswordClientSaveRepository {
	return &SecretUsernamePasswordClientSaveRepository{db: db}
}

// Save сохраняет или обновляет секрет с логином и паролем клиента в базе данных.
//
// Если возникает конфликт по secret_name, то выполняется обновление соответствующих полей.
// Иначе создаётся новая запись.
//
// Параметры:
//   - ctx: контекст для управления временем выполнения запроса и отмены.
//   - secret: структура с данными секрета (логин и пароль).
//
// Возвращает ошибку в случае сбоя выполнения SQL-запроса.
func (r *SecretUsernamePasswordClientSaveRepository) Save(
	ctx context.Context,
	secret models.SecretUsernamePasswordClient,
) error {
	query := `
		INSERT INTO secret_username_password (secret_name, username, password, meta, updated_at)
		VALUES (:secret_name, :username, :password, :meta, :updated_at)
		ON CONFLICT(secret_name) DO UPDATE SET
			username = excluded.username,
			password = excluded.password,
			meta = excluded.meta,
			updated_at = excluded.updated_at;
	`

	_, err := r.db.NamedExecContext(ctx, query, secret)
	return err
}

// SecretUsernamePasswordClientListRepository предоставляет методы для получения секретов типа "логин-пароль" клиента из базы данных.
type SecretUsernamePasswordClientListRepository struct {
	db *sqlx.DB
}

// NewSecretUsernamePasswordClientListRepository создаёт новый экземпляр SecretUsernamePasswordClientListRepository.
func NewSecretUsernamePasswordClientListRepository(db *sqlx.DB) *SecretUsernamePasswordClientListRepository {
	return &SecretUsernamePasswordClientListRepository{db: db}
}

// List возвращает все секреты с логином и паролем клиента из базы данных.
//
// Параметры:
//   - ctx: контекст для управления временем выполнения и отмены операции.
//
// Возвращает:
//   - []models.SecretUsernamePasswordClient: список всех найденных секретов
//   - error: ошибка, если операция завершилась неудачей
func (r *SecretUsernamePasswordClientListRepository) List(ctx context.Context) ([]models.SecretUsernamePasswordClient, error) {
	query := `
		SELECT secret_name, username, password, meta, updated_at
		FROM secret_username_password;
	`

	var secrets []models.SecretUsernamePasswordClient
	if err := r.db.SelectContext(ctx, &secrets, query); err != nil {
		return nil, err
	}

	return secrets, nil
}
