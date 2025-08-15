package tx

import (
	"context"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	_ "modernc.org/sqlite"
)

func setupDB(t *testing.T) *sqlx.DB {
	db, err := sqlx.Open("sqlite", ":memory:")
	require.NoError(t, err)
	return db
}

func TestTx_SetAndGet(t *testing.T) {
	db := setupDB(t)
	defer db.Close()

	txManager := New(db)

	ctx := context.Background()

	// Создаем транзакцию через Set
	ctxWithTx, err := txManager.Set(ctx)
	require.NoError(t, err)

	// Получаем транзакцию через Get
	gotTx := txManager.Get(ctxWithTx)
	require.NotNil(t, gotTx, "транзакция должна быть в контексте")
}

func TestTx_Get_NoTx(t *testing.T) {
	db := setupDB(t)
	defer db.Close()

	txManager := New(db)

	ctx := context.Background()

	// Попытка получить транзакцию из контекста, где её нет
	gotTx := txManager.Get(ctx)
	assert.Nil(t, gotTx, "транзакции не должно быть в контексте")
}

func TestTx_Set_Error(t *testing.T) {
	// Передаем закрытую БД для генерации ошибки BeginTxx
	db := setupDB(t)
	db.Close() // Закрываем для эмуляции ошибки

	txManager := New(db)
	ctx := context.Background()

	_, err := txManager.Set(ctx)
	assert.Error(t, err, "должна быть ошибка при создании транзакции на закрытой БД")
}
