package app

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	_ "modernc.org/sqlite"
)

func TestNewCommand(t *testing.T) {
	cmd := NewCommand()

	assert.Equal(t, "server", cmd.Use)
	assert.Equal(t, "Запускает HTTP сервер GophKeeper", cmd.Short)

	flags := cmd.Flags()

	serverURL, err := flags.GetString("server-url")
	assert.NoError(t, err)
	assert.Equal(t, ":8080", serverURL) // исправлено

	dbDriver, err := flags.GetString("database-driver")
	assert.NoError(t, err)
	assert.Equal(t, "sqlite", dbDriver) // исправлено

	dbDSN, err := flags.GetString("database-dsn")
	assert.NoError(t, err)
	assert.Equal(t, "server.db", dbDSN) // исправлено

	maxOpenConns, err := flags.GetInt("database-max-open-conns")
	assert.NoError(t, err)
	assert.Equal(t, 10, maxOpenConns)

	maxIdleConns, err := flags.GetInt("database-max-idle-conns")
	assert.NoError(t, err)
	assert.Equal(t, 5, maxIdleConns)

	connMaxLifetime, err := flags.GetDuration("database-conn-max-lifetime")
	assert.NoError(t, err)
	assert.Equal(t, time.Hour, connMaxLifetime)

	migrationsDir, err := flags.GetString("migrations-dir")
	assert.NoError(t, err)
	assert.Equal(t, "migrations", migrationsDir)

	jwtSecret, err := flags.GetString("jwt-secret")
	assert.NoError(t, err)
	assert.Equal(t, "secret", jwtSecret)

	jwtExp, err := flags.GetDuration("jwt-exp")
	assert.NoError(t, err)
	assert.Equal(t, 24*time.Hour, jwtExp)
}

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
		errCh <- runHTTP(
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
