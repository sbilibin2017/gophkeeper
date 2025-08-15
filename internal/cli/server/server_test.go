package server

import (
	"context"
	"testing"
	"time"

	"github.com/sbilibin2017/gophkeeper/internal/configs/address"
	"github.com/stretchr/testify/assert"
)

func TestNewCommand_RunE_HTTPRunnerCalled(t *testing.T) {
	called := false

	mockRunner := func(
		ctx context.Context,
		serverURL string,
		databaseDSN string,
		databaseMigrationsDir string,
		jwtSecretKey string,
		jwtExp time.Duration,
	) error {
		called = true

		// Используем нормализованный адрес
		assert.Equal(t, "0.0.0.0:8080", serverURL)
		assert.Equal(t, "gophkeeper.db", databaseDSN)
		assert.Equal(t, "migrations", databaseMigrationsDir)
		assert.Equal(t, "secret", jwtSecretKey)
		assert.Equal(t, time.Hour, jwtExp)

		return nil
	}

	cmd := NewCommand(mockRunner)
	assert.NotNil(t, cmd)

	err := cmd.Execute()
	assert.NoError(t, err)
	assert.True(t, called, "httpRunner should be called")
}

func TestNewCommand_RunE_CustomFlags(t *testing.T) {
	called := false

	mockRunner := func(
		ctx context.Context,
		serverURL string,
		databaseDSN string,
		databaseMigrationsDir string,
		jwtSecretKey string,
		jwtExp time.Duration,
	) error {
		called = true

		assert.Equal(t, "127.0.0.1:9000", serverURL) // <-- без http://
		assert.Equal(t, "custom.db", databaseDSN)
		assert.Equal(t, "custom_migrations", databaseMigrationsDir)
		assert.Equal(t, "mysecret", jwtSecretKey)
		assert.Equal(t, 2*time.Hour, jwtExp)

		return nil
	}

	cmd := NewCommand(mockRunner)
	cmd.SetArgs([]string{
		"--address", "http://127.0.0.1:9000",
		"--database-dsn", "custom.db",
		"--migrations-dir", "custom_migrations",
		"--jwt-secret", "mysecret",
		"--jwt-exp", "2h",
	})

	err := cmd.Execute()
	assert.NoError(t, err)
	assert.True(t, called, "httpRunner should be called")
}

func TestNewCommand_RunE_UnsupportedScheme_Error(t *testing.T) {
	called := false

	mockRunner := func(
		ctx context.Context,
		serverURL string,
		databaseDSN string,
		databaseMigrationsDir string,
		jwtSecretKey string,
		jwtExp time.Duration,
	) error {
		called = true
		return nil
	}

	cmd := NewCommand(mockRunner)

	// Используем явно неподдерживаемую схему
	cmd.SetArgs([]string{"--address", "unsupported://example.com:1234"})

	err := cmd.Execute()

	assert.Error(t, err)
	assert.Equal(t, address.ErrUnsupportedScheme, err)
	assert.False(t, called, "httpRunner should NOT be called for unsupported scheme")
}
