package apps

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	_ "modernc.org/sqlite"
)

func TestRunServerHTTP_RunOnly(t *testing.T) {
	// создаём временную директорию для миграций
	migrationsDir := "./migrations_test"
	os.MkdirAll(migrationsDir, 0755)
	defer os.RemoveAll(migrationsDir)

	// создаём минимальную миграцию для sqlite
	migrationFile := filepath.Join(migrationsDir, "00001_create_users_table.sql")
	sql := `-- +goose Up
CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT NOT NULL,
    password TEXT NOT NULL
);

-- +goose Down
DROP TABLE users;
`
	os.WriteFile(migrationFile, []byte(sql), 0644)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	serverURL := "localhost:0" // 0 — выбрать свободный порт
	databaseDriver := "sqlite"
	databaseDSN := ":memory:"

	jwtSecret := "testsecret"
	jwtExp := time.Minute * 5

	errCh := make(chan error, 1)

	// запускаем сервер в отдельной горутине
	go func() {
		errCh <- RunServerHTTP(
			ctx,
			serverURL,
			databaseDriver,
			databaseDSN,
			5, 5, time.Minute,
			migrationsDir,
			jwtSecret,
			jwtExp,
		)
	}()

	// ждём 1 секунду, чтобы сервер стартовал (можно убрать, если используешь сигналы)
	time.Sleep(time.Second * 1)

	// отменяем контекст для graceful shutdown
	cancel()

	// ждём завершения сервера и проверяем ошибки
	err := <-errCh
	assert.NoError(t, err, "сервер не должен возвращать ошибку при запуске")
}
