package apps

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRunHTTP_StartAndShutdown(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Используем SQLite в памяти
	dsn := ":memory:"
	migrationsDir := "" // пропускаем миграции для простоты
	serverURL := "127.0.0.1:8085"
	jwtSecret := "testsecret"
	jwtExp := time.Hour

	// Запускаем сервер в горутине
	go func() {
		err := RunServerHTTP(ctx, serverURL, dsn, migrationsDir, jwtSecret, jwtExp)
		assert.NoError(t, err)
	}()

	// Ждем немного, чтобы сервер поднялся
	time.Sleep(500 * time.Millisecond)

	// Останавливаем сервер
	cancel()
	time.Sleep(200 * time.Millisecond)
}
