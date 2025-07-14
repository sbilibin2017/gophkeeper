package repositories

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/sbilibin2017/gophkeeper/internal/models"
)

// SecretBankCardClientSaveRepository обеспечивает сохранение секретных данных банковской карты клиента в базе данных.
type SecretBankCardClientSaveRepository struct {
	db *sqlx.DB
}

// NewSecretBankCardClientSaveRepository создаёт новый экземпляр SecretBankCardClientSaveRepository.
func NewSecretBankCardClientSaveRepository(db *sqlx.DB) *SecretBankCardClientSaveRepository {
	return &SecretBankCardClientSaveRepository{db: db}
}

// Save сохраняет или обновляет запись о секретной банковской карте клиента в базе данных.
//
// Если запись с таким же secret_name и owner уже существует, она будет обновлена.
// В противном случае — вставлена новая запись.
//
// Параметры:
//   - ctx: контекст для управления таймаутами и отменой запроса.
//   - secret: структура с данными секретной банковской карты клиента.
//
// Возвращает ошибку, если выполнение запроса в базу данных завершилось неудачно.
func (r *SecretBankCardClientSaveRepository) Save(
	ctx context.Context,
	secret models.SecretBankCardClient,
) error {
	query := `
		INSERT INTO secret_bank_card (secret_name, owner, number, exp, cvv, meta, updated_at)
		VALUES (:secret_name, :owner, :number, :exp, :cvv, :meta, :updated_at)
		ON CONFLICT(secret_name, owner) DO UPDATE SET
			number = excluded.number,
			exp = excluded.exp,
			cvv = excluded.cvv,
			meta = excluded.meta,
			updated_at = excluded.updated_at;
	`

	_, err := r.db.NamedExecContext(ctx, query, secret)
	return err
}

// SecretBankCardClientListRepository обеспечивает чтение всех секретных банковских карт клиента из базы данных.
type SecretBankCardClientListRepository struct {
	db *sqlx.DB
}

// NewSecretBankCardClientListRepository создаёт новый экземпляр SecretBankCardClientListRepository.
func NewSecretBankCardClientListRepository(db *sqlx.DB) *SecretBankCardClientListRepository {
	return &SecretBankCardClientListRepository{db: db}
}

// List возвращает все записи секретных банковских карт клиента из базы данных.
//
// Параметры:
//   - ctx: контекст для управления таймаутами и отменой запроса.
//
// Возвращает:
//   - []models.SecretBankCardClient: список всех найденных записей
//   - error: ошибка, если чтение не удалось
func (r *SecretBankCardClientListRepository) List(ctx context.Context) ([]models.SecretBankCardClient, error) {
	query := `
		SELECT secret_name, owner, number, exp, cvv, meta, updated_at
		FROM secret_bank_card_client;
	`

	var secrets []models.SecretBankCardClient
	if err := r.db.SelectContext(ctx, &secrets, query); err != nil {
		return nil, err
	}

	return secrets, nil
}
