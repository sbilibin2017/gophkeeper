package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewDB_Success(t *testing.T) {
	// Используем in-memory SQLite базу для теста
	db, err := NewDB(":memory:")
	assert.NoError(t, err)
	assert.NotNil(t, db)

	// Закрываем базу после теста
	if db != nil {
		_ = db.Close()
	}
}

func TestNewDB_InvalidPath(t *testing.T) {
	// Попытка открыть базу по некорректному пути должна вернуть ошибку
	db, err := NewDB("/invalid/path/to/db.sqlite")
	assert.Error(t, err)
	assert.Nil(t, db)
}
