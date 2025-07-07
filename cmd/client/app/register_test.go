package app

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"

	"github.com/sbilibin2017/gophkeeper/internal/configs"
)

func TestParseRegisterFlags_ValidHTTP(t *testing.T) {
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

	config, creds, err := parseRegisterFlags(cmd)
	require.NoError(t, err)
	require.NotNil(t, config)
	require.NotNil(t, creds)

	require.Equal(t, "testuser", creds.Username)
	require.Equal(t, "testpass", creds.Password)
	require.NotNil(t, config.HTTPClient)
	require.Nil(t, config.GRPCClient)
	require.NotNil(t, config.HMACEncoder)
}

func TestParseRegisterFlags_InvalidURL(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().String("username", "", "")
	cmd.Flags().String("password", "", "")
	cmd.Flags().String("server-url", "", "")

	require.NoError(t, cmd.Flags().Set("username", "testuser"))
	require.NoError(t, cmd.Flags().Set("password", "testpass"))
	require.NoError(t, cmd.Flags().Set("server-url", "ftp://invalid"))

	_, _, err := parseRegisterFlags(cmd)
	require.Error(t, err)
	require.Contains(t, err.Error(), "unsupported URL scheme")
}

func TestNewRegisterService_HTTP(t *testing.T) {
	cfg := &configs.ClientConfig{}
	// Используем опцию WithClient с HTTP URL для инициализации HTTP клиента
	err := configs.WithClient("http://localhost")(cfg)
	require.NoError(t, err)
	require.NotNil(t, cfg.HTTPClient)
	require.Nil(t, cfg.GRPCClient)

	svc, err := newRegisterService(cfg)
	require.NoError(t, err)
	require.NotNil(t, svc)
}

func TestNewRegisterService_GRPC(t *testing.T) {
	cfg := &configs.ClientConfig{}
	// Для теста gRPC создадим невалидное соединение (Dial к localhost), но главное - наличие grpc.ClientConn
	err := configs.WithClient("grpc://localhost:12345")(cfg)
	require.NoError(t, err)
	require.NotNil(t, cfg.GRPCClient)
	require.Nil(t, cfg.HTTPClient)

	svc, err := newRegisterService(cfg)
	require.NoError(t, err)
	require.NotNil(t, svc)

	// Закрываем соединение после теста
	_ = cfg.GRPCClient.Close()
}

func TestNewRegisterService_UnsupportedScheme(t *testing.T) {
	cfg := &configs.ClientConfig{}
	svc, err := newRegisterService(cfg)
	require.Error(t, err)
	require.Nil(t, svc)
}
