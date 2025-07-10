package options

import (
	"os"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewOptions_NewClientConfigReturnsError(t *testing.T) {
	_, err := NewOptions(
		WithToken("token"),
		WithServerURL("://bad-url"),
	)

	require.Error(t, err)
}

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

func TestRegisterTokenFlag(t *testing.T) {
	var token string
	cmd := &cobra.Command{}
	cmd = RegisterTokenFlag(cmd, &token)

	flag := cmd.Flags().Lookup("token")
	assert.NotNil(t, flag)
	assert.Equal(t, "", flag.DefValue)
	assert.Equal(t, "Токен авторизации (можно задать через GOPHKEEPER_TOKEN)", flag.Usage)

	err := cmd.Flags().Set("token", "mytoken")
	assert.NoError(t, err)
	assert.Equal(t, "mytoken", token)
}

func TestRegisterServerURLFlag(t *testing.T) {
	var serverURL string
	cmd := &cobra.Command{}
	cmd = RegisterServerURLFlag(cmd, &serverURL)

	flag := cmd.Flags().Lookup("server-url")
	assert.NotNil(t, flag)
	assert.Equal(t, "", flag.DefValue)
	assert.Equal(t, "URL сервера (можно задать через GOPHKEEPER_SERVER_URL)", flag.Usage)

	err := cmd.Flags().Set("server-url", "https://example.com")
	assert.NoError(t, err)
	assert.Equal(t, "https://example.com", serverURL)
}

func TestRegisterInteractiveFlag(t *testing.T) {
	var interactive bool
	cmd := &cobra.Command{}
	cmd = RegisterInteractiveFlag(cmd, &interactive)

	flag := cmd.Flags().Lookup("interactive")
	assert.NotNil(t, flag)
	assert.Equal(t, "false", flag.DefValue)
	assert.Equal(t, "Включить интерактивный режим ввода", flag.Usage)

	err := cmd.Flags().Set("interactive", "true")
	assert.NoError(t, err)
	assert.True(t, interactive)
}
