package app

import (
	"testing"

	"github.com/sbilibin2017/gophkeeper/internal/configs"
	"github.com/stretchr/testify/require"
)

func TestNewClientConfig_HTTP(t *testing.T) {
	cfg, err := configs.NewClientConfig(
		configs.WithClient("http://localhost"),
		configs.WithHMACEncoder("secret"),
	)
	require.NoError(t, err)
	require.NotNil(t, cfg.HTTPClient)
	require.Nil(t, cfg.GRPCClient)
	require.NotNil(t, cfg.HMACEncoder)
}

func TestNewClientConfig_GRPC(t *testing.T) {
	cfg, err := configs.NewClientConfig(
		configs.WithClient("grpc://localhost:50051"),
	)
	require.NoError(t, err)
	require.NotNil(t, cfg.GRPCClient)
	require.Nil(t, cfg.HTTPClient)
}

func TestNewClientConfig_UnsupportedScheme(t *testing.T) {
	_, err := configs.NewClientConfig(
		configs.WithClient("ftp://invalid"),
	)
	require.Error(t, err)
	require.Contains(t, err.Error(), "unsupported URL scheme")
}
