package commands

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewServerCommand_RunHTTPCalled(t *testing.T) {
	var (
		called            bool
		capturedURL       string
		capturedDSN       string
		capturedJWTSecret string
		capturedJWTExp    time.Duration
	)

	// Подставляем заглушку вместо реального runHTTP
	runHTTPStub := func(
		ctx context.Context,
		serverURL string,
		databaseDriver string,
		databaseDSN string,
		databaseMaxOpenConns int,
		databaseMaxIdleConns int,
		databaseConnMaxLifetime time.Duration,
		migrationsDir string,
		jwtSecret string,
		jwtExp time.Duration,
	) error {
		called = true
		capturedURL = serverURL
		capturedDSN = databaseDSN
		capturedJWTSecret = jwtSecret
		capturedJWTExp = jwtExp
		return nil
	}

	cmd := NewServerCommand(runHTTPStub)

	// Симулируем запуск команды с флагами
	cmd.SetArgs([]string{
		"--server-url", "localhost:9090",
		"--database-dsn", "test-dsn",
		"--jwt-secret", "supersecret",
		"--jwt-exp", "2h",
	})

	err := cmd.Execute()
	assert.NoError(t, err, "команда не должна возвращать ошибку")
	assert.True(t, called, "runHTTP должен быть вызван")
	assert.Equal(t, "http://localhost:9090", capturedURL) // <- схема http:// добавлена автоматически
	assert.Equal(t, "test-dsn", capturedDSN)
	assert.Equal(t, "supersecret", capturedJWTSecret)
	assert.Equal(t, 2*time.Hour, capturedJWTExp)
}
