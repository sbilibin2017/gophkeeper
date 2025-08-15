package tx

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
)

// Tx управляет транзакциями базы данных.
type Tx struct {
	db *sqlx.DB
}

// txKey используется как ключ для хранения транзакции в контексте.
type txKey struct{}

// New создает новый объект Tx для работы с транзакциями на указанной базе данных.
//
// Пример использования:
//
//	txManager := tx.New(db)
func New(db *sqlx.DB) *Tx {
	return &Tx{
		db: db,
	}
}

// Set начинает новую транзакцию и сохраняет её в контексте.
//
// Возвращает новый контекст с транзакцией или ошибку, если транзакцию создать не удалось.
func (t *Tx) Set(ctx context.Context) (context.Context, error) {
	tx, err := t.db.BeginTxx(ctx, &sql.TxOptions{})
	if err != nil {
		return ctx, err
	}
	return context.WithValue(ctx, txKey{}, tx), nil
}

// Get извлекает текущую транзакцию из контекста.
//
// Если транзакция отсутствует, возвращается nil.
func (t *Tx) Get(ctx context.Context) *sqlx.Tx {
	tx, _ := ctx.Value(txKey{}).(*sqlx.Tx)
	return tx
}
