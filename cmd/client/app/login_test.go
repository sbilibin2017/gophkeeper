package app

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"

	"github.com/sbilibin2017/gophkeeper/internal/configs"
)

func TestParseLoginFlags_ValidHTTP(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().String("username", "", "")
	cmd.Flags().String("password", "", "")
	cmd.Flags().String("server-url", "", "")
	cmd.Flags().String("hmac-key", "", "")
	cmd.Flags().String("rsa-public-key", "", "")

	require.NoError(t, cmd.Flags().Set("username", "testuser"))
	require.NoError(t, cmd.Flags().Set("password", "testpass"))
	require.NoError(t, cmd.Flags().Set("server-url", "http://localhost"))
	require.NoError(t, cmd.Flags().Set("hmac-key", "mysecret"))

	config, creds, err := parseLoginFlags(cmd)
	require.NoError(t, err)
	require.NotNil(t, config)
	require.NotNil(t, creds)

	require.Equal(t, "testuser", creds.Username)
	require.Equal(t, "testpass", creds.Password)
	require.NotNil(t, config.HTTPClient)
	require.Nil(t, config.GRPCClient)
	require.NotNil(t, config.HMACEncoder)
}

func TestParseLoginFlags_InvalidURL(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().String("username", "", "")
	cmd.Flags().String("password", "", "")
	cmd.Flags().String("server-url", "", "")

	require.NoError(t, cmd.Flags().Set("username", "testuser"))
	require.NoError(t, cmd.Flags().Set("password", "testpass"))
	require.NoError(t, cmd.Flags().Set("server-url", "ftp://invalid"))

	_, _, err := parseLoginFlags(cmd)
	require.Error(t, err)
	require.Contains(t, err.Error(), "unsupported URL scheme")
}

func TestNewLoginService_HTTP(t *testing.T) {
	cfg := &configs.ClientConfig{}
	err := configs.WithClient("http://localhost")(cfg)
	require.NoError(t, err)
	require.NotNil(t, cfg.HTTPClient)
	require.Nil(t, cfg.GRPCClient)

	svc, err := newLoginService(cfg)
	require.NoError(t, err)
	require.NotNil(t, svc)

}

func TestNewLoginService_GRPC(t *testing.T) {
	cfg := &configs.ClientConfig{}
	err := configs.WithClient("grpc://localhost:12345")(cfg)
	require.NoError(t, err)
	require.NotNil(t, cfg.GRPCClient)
	require.Nil(t, cfg.HTTPClient)

	svc, err := newLoginService(cfg)
	require.NoError(t, err)
	require.NotNil(t, svc)

	_ = cfg.GRPCClient.Close()
}

func TestNewLoginService_UnsupportedScheme(t *testing.T) {
	cfg := &configs.ClientConfig{}
	svc, err := newLoginService(cfg)
	require.Error(t, err)
	require.Nil(t, svc)
}
