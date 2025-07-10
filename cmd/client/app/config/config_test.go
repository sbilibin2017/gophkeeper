package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewConfig_WithValidHTTP(t *testing.T) {
	cfg, err := NewConfig(
		WithToken("test-token"),
		WithServerURL("https://example.com"),
	)
	require.NoError(t, err)
	require.NotNil(t, cfg)
	assert.Equal(t, "test-token", cfg.Token)
	assert.Equal(t, "https://example.com", cfg.ServerURL)
	assert.NotNil(t, cfg.ClientConfig)
}

func TestNewConfig_WithValidGRPC(t *testing.T) {
	cfg, err := NewConfig(
		WithToken("grpc-token"),
		WithServerURL("grpc://localhost:50051"),
	)
	require.NoError(t, err)
	require.NotNil(t, cfg)
	assert.Equal(t, "grpc-token", cfg.Token)
	assert.Equal(t, "grpc://localhost:50051", cfg.ServerURL)
	assert.NotNil(t, cfg.ClientConfig)
}

func TestNewConfig_UsesEnvVars(t *testing.T) {
	os.Setenv("GOPHKEEPER_TOKEN", "env-token")
	os.Setenv("GOPHKEEPER_SERVER_URL", "https://env-server.com")
	t.Cleanup(func() {
		os.Unsetenv("GOPHKEEPER_TOKEN")
		os.Unsetenv("GOPHKEEPER_SERVER_URL")
	})

	cfg, err := NewConfig()
	require.NoError(t, err)
	require.NotNil(t, cfg)
	assert.Equal(t, "env-token", cfg.Token)
	assert.Equal(t, "https://env-server.com", cfg.ServerURL)
	assert.NotNil(t, cfg.ClientConfig)
}

func TestNewConfig_InvalidURL(t *testing.T) {
	_, err := NewConfig(
		WithToken("token"),
		WithServerURL(":::bad-url"),
	)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid server URL")
}

func TestNewConfig_UnsupportedScheme(t *testing.T) {
	_, err := NewConfig(
		WithToken("token"),
		WithServerURL("ftp://example.com"),
	)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported URL scheme")
}
