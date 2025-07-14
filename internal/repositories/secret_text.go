package repositories

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/sbilibin2017/gophkeeper/internal/models"
)

// SecretTextClientSaveRepository отвечает за сохранение текстовых секретов клиента в базе данных.
type SecretTextClientSaveRepository struct {
	db *sqlx.DB
}

// NewSecretTextClientSaveRepository создаёт новый репозиторий для работы с текстовыми секретами клиента.
func NewSecretTextClientSaveRepository(db *sqlx.DB) *SecretTextClientSaveRepository {
	return &SecretTextClientSaveRepository{db: db}
}

// Save сохраняет или обновляет текстовый секрет клиента в базе данных.
//
// При наличии конфликта по secret_name происходит обновление соответствующих полей.
// Иначе создаётся новая запись.
//
// Параметры:
//   - ctx: контекст для управления временем выполнения запроса и отмены.
//   - secret: структура с данными текстового секрета клиента.
//
// Возвращает ошибку в случае неудачного выполнения запроса к базе данных.
func (r *SecretTextClientSaveRepository) Save(
	ctx context.Context,
	secret models.SecretTextClient,
) error {
	query := `
		INSERT INTO secret_text (secret_name, content, meta, updated_at)
		VALUES (:secret_name, :content, :meta, :updated_at)
		ON CONFLICT(secret_name) DO UPDATE SET
			content = excluded.content,
			meta = excluded.meta,
			updated_at = excluded.updated_at;
	`

	_, err := r.db.NamedExecContext(ctx, query, secret)
	return err
}
