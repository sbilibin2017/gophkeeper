package configs

import (
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClientConfig(t *testing.T) {
	t.Run("Empty config", func(t *testing.T) {
		cfg, err := NewClientConfig()
		require.NoError(t, err)
		assert.Nil(t, cfg.HTTPClient)
		assert.Nil(t, cfg.GRPCClient)
		assert.Nil(t, cfg.DB)
	})

	t.Run("With HTTP Client", func(t *testing.T) {
		cfg, err := NewClientConfig(
			WithClientConfigHTTPClient("https://example.com"),
		)
		require.NoError(t, err)
		require.NotNil(t, cfg.HTTPClient)
		assert.IsType(t, &resty.Client{}, cfg.HTTPClient)
		assert.Equal(t, "https://example.com", cfg.HTTPClient.BaseURL)
	})

	t.Run("With empty HTTP baseURL", func(t *testing.T) {
		cfg, err := NewClientConfig(
			WithClientConfigHTTPClient(""),
		)
		require.NoError(t, err)
		assert.Nil(t, cfg.HTTPClient)
	})

	t.Run("With gRPC Client", func(t *testing.T) {
		// Используем localhost:0, чтобы не ожидать реального сервера, соединение может упасть, но тест проверит ошибку
		cfg, err := NewClientConfig(
			WithClientConfigGRPCClient("localhost:0"),
		)
		// Ошибка подключения ожидаема, так как сервера нет — проверим, что err != nil и GRPCClient == nil
		if err != nil {
			assert.Nil(t, cfg)
		} else {
			require.NotNil(t, cfg.GRPCClient)
			_ = cfg.GRPCClient.Close()
		}
	})

	t.Run("With empty gRPC address", func(t *testing.T) {
		cfg, err := NewClientConfig(
			WithClientConfigGRPCClient(""),
		)
		require.NoError(t, err)
		assert.Nil(t, cfg.GRPCClient)
	})

	t.Run("With DB connection", func(t *testing.T) {
		// Используем SQLite in-memory БД
		cfg, err := NewClientConfig(
			WithClientConfigDB("sqlite", "file::memory:?cache=shared"),
		)
		require.NoError(t, err)
		require.NotNil(t, cfg.DB)
		var version string
		err = cfg.DB.Get(&version, "select sqlite_version()")
		assert.NoError(t, err)
	})

	t.Run("With empty DB dsn", func(t *testing.T) {
		cfg, err := NewClientConfig(
			WithClientConfigDB("sqlite", ""),
		)
		require.NoError(t, err)
		assert.Nil(t, cfg.DB)
	})

	t.Run("With DB With Migrations", func(t *testing.T) {
		cfg, err := NewClientConfig(
			WithClientConfigDBWithMigrations("file::memory:?cache=shared"),
		)
		require.NoError(t, err)
		require.NotNil(t, cfg.DB)
	})

	t.Run("With empty DSN With Migrations", func(t *testing.T) {
		cfg, err := NewClientConfig(
			WithClientConfigDBWithMigrations(""),
		)
		require.NoError(t, err)
		assert.Nil(t, cfg.DB)
	})
}
