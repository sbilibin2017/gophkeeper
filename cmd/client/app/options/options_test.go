package options

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewOptions_WithValidHTTP(t *testing.T) {
	cfg, err := NewOptions(
		WithToken("test-token"),
		WithServerURL("https://example.com"),
	)
	require.NoError(t, err)
	require.NotNil(t, cfg)
	assert.Equal(t, "test-token", cfg.Token)
	assert.Equal(t, "https://example.com", cfg.ServerURL)
	assert.NotNil(t, cfg.ClientConfig)
	assert.NotNil(t, cfg.ClientConfig.HTTPClient)
	assert.Nil(t, cfg.ClientConfig.GRPCClient)
}

func TestNewOptions_WithValidGRPC(t *testing.T) {
	cfg, err := NewOptions(
		WithToken("grpc-token"),
		WithServerURL("grpc://localhost:50051"),
	)
	require.NoError(t, err)
	require.NotNil(t, cfg)
	assert.Equal(t, "grpc-token", cfg.Token)
	assert.Equal(t, "grpc://localhost:50051", cfg.ServerURL)
	assert.NotNil(t, cfg.ClientConfig)
	assert.Nil(t, cfg.ClientConfig.HTTPClient)
	// GRPCClient может быть nil, если не удаётся подключиться
}

func TestNewOptions_UsesEnvVars(t *testing.T) {
	os.Setenv("GOPHKEEPER_TOKEN", "env-token")
	os.Setenv("GOPHKEEPER_SERVER_URL", "https://env-server.com")
	t.Cleanup(func() {
		os.Unsetenv("GOPHKEEPER_TOKEN")
		os.Unsetenv("GOPHKEEPER_SERVER_URL")
	})

	cfg, err := NewOptions()
	require.NoError(t, err)
	require.NotNil(t, cfg)
	assert.Equal(t, "env-token", cfg.Token)
	assert.Equal(t, "https://env-server.com", cfg.ServerURL)
	assert.NotNil(t, cfg.ClientConfig)
}

func TestNewOptions_InvalidURL(t *testing.T) {
	_, err := NewOptions(
		WithToken("token"),
		WithServerURL(":::bad-url"),
	)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid server URL")
}

func TestNewOptions_UnsupportedScheme(t *testing.T) {
	_, err := NewOptions(
		WithToken("token"),
		WithServerURL("ftp://example.com"),
	)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported URL scheme")
}

func TestSetToken(t *testing.T) {
	err := SetToken("new-token")
	require.NoError(t, err)
	assert.Equal(t, "new-token", os.Getenv("GOPHKEEPER_TOKEN"))

	err = SetToken("")
	require.Error(t, err)
}

func TestSetServerURL(t *testing.T) {
	err := SetServerURL("https://example.com")
	require.NoError(t, err)
	assert.Equal(t, "https://example.com", os.Getenv("GOPHKEEPER_SERVER_URL"))

	err = SetServerURL("")
	require.Error(t, err)
}
