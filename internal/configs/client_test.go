package configs

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWithHTTPClient(t *testing.T) {
	serverURL := "https://example.com/api"
	cfg, err := NewClientConfig(WithHTTPClient(serverURL))
	require.NoError(t, err)
	require.NotNil(t, cfg.HTTPClient)
	assert.Equal(t, serverURL, cfg.HTTPClient.BaseURL)
}

func TestWithGRPCClient(t *testing.T) {
	serverURL := "grpc://localhost:50051"
	cfg, err := NewClientConfig(WithGRPCClient(serverURL))
	require.NoError(t, err)
	require.NotNil(t, cfg.GRPCClient)

	parsed, err := url.Parse(serverURL)
	require.NoError(t, err)
	assert.Contains(t, cfg.GRPCClient.Target(), parsed.Host)

	cfg.GRPCClient.Close()
}
