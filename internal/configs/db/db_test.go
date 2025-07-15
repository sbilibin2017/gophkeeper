package db

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewDB(t *testing.T) {
	// Успешное подключение к SQLite в памяти
	db, err := NewDB("sqlite", ":memory:")
	assert.NoError(t, err)
	assert.NotNil(t, db)
	db.Close()

	// Ошибка при неизвестном драйвере
	db, err = NewDB("unknown_driver", "dsn")
	assert.Error(t, err)
	assert.Nil(t, db)
}

func TestRunMigrations(t *testing.T) {
	db, err := NewDB("sqlite", ":memory:")
	assert.NoError(t, err)
	defer db.Close()

	// Создаем временную директорию для миграций
	dir, err := os.MkdirTemp("", "migrations")
	assert.NoError(t, err)
	defer os.RemoveAll(dir) // Очистка после теста

	// Создаем файл миграции в этой директории
	migrationFile := filepath.Join(dir, "0001_init.sql")
	content := `-- +goose Up
CREATE TABLE test_table(id INTEGER PRIMARY KEY);

-- +goose Down
DROP TABLE test_table;
`
	err = os.WriteFile(migrationFile, []byte(content), 0644)
	assert.NoError(t, err)

	// Запускаем миграции из временной директории
	err = RunMigrations(db, "sqlite", dir)
	assert.NoError(t, err)

	// Можно проверить, что таблица создана (опционально)
	var count int
	err = db.Get(&count, "SELECT count(*) FROM sqlite_master WHERE type='table' AND name='test_table'")
	assert.NoError(t, err)
	assert.Equal(t, 1, count)

	// Тест ошибки установки диалекта
	err = RunMigrations(db, "unknown_driver", dir)
	assert.Error(t, err)
}
