package configs

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWithToken(t *testing.T) {
	cfg, err := NewClientConfig(WithToken("test-token"))
	require.NoError(t, err)
	assert.Equal(t, "test-token", cfg.Token)

	cfg, err = NewClientConfig(WithToken(""))
	require.NoError(t, err)
	assert.Equal(t, "", cfg.Token)
}

func TestWithServerURL(t *testing.T) {
	cfg, err := NewClientConfig(WithServerURL("https://example.com"))
	require.NoError(t, err)
	assert.Equal(t, "https://example.com", cfg.ServerURL)

	cfg, err = NewClientConfig(WithServerURL(""))
	require.NoError(t, err)
	assert.Equal(t, "", cfg.ServerURL)
}

func TestWithDB(t *testing.T) {
	_, err := NewClientConfig(WithDB("/invalid/path/to/db.sqlite"))
	require.Error(t, err)

	cfg, err := NewClientConfig(WithDB(":memory:"))
	require.NoError(t, err)
	require.NotNil(t, cfg.DB)

	err = cfg.DB.Close()
	require.NoError(t, err)
}

func TestWithHTTPClient(t *testing.T) {
	cfg, err := NewClientConfig(WithHTTPClient("https://example.com"))
	require.NoError(t, err)
	require.NotNil(t, cfg.HTTPClient)
	assert.Equal(t, "https://example.com", cfg.HTTPClient.BaseURL)

	_, err = NewClientConfig(WithHTTPClient("://bad-url"))
	require.Error(t, err)

	_, err = NewClientConfig(WithHTTPClient("http://"))
	require.Error(t, err)
}

func TestWithGRPCClient(t *testing.T) {
	_, err := NewClientConfig(WithGRPCClient("://bad-url"))
	require.Error(t, err)

	_, err = NewClientConfig(WithGRPCClient("http://"))
	require.Error(t, err)

	cfg, err := NewClientConfig(WithGRPCClient("grpc://localhost:12345"))
	require.NotNil(t, cfg)
	require.Nil(t, err)
}
