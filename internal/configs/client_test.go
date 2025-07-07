package configs

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWithHMACEncoder(t *testing.T) {
	key := "mysecret"
	cfg, err := NewClientConfig(WithHMACEncoder(key))
	require.NoError(t, err)
	require.NotNil(t, cfg)
	require.NotNil(t, cfg.hmacEncoder)

	data := []byte("test")
	mac := cfg.hmacEncoder(data)
	require.NotEmpty(t, mac)
}

func TestWithHMACEncoder_EmptyKey(t *testing.T) {
	_, err := NewClientConfig(WithHMACEncoder(""))
	require.Error(t, err)
	require.Contains(t, err.Error(), "HMAC key cannot be empty")
}

func TestWithRSAEncoder_InvalidPath(t *testing.T) {
	_, err := NewClientConfig(WithRSAEncoder("no-such-file.pem"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to read RSA public key file")
}

func TestWithRSAEncoder_ValidKey(t *testing.T) {
	const pemData = `-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAwXTAvZiw5eE0LGB79u0K
7M1EnuW1rZOmD5sKeac1TIDrbi7MeME8ONxWHP8bHD+nnhcX3F0PiI98bhQhVctN
M5EOhBBv1KhKNflRMgJzvVGuqJAGxUv5C8sPa2F4N8A9HYIHRtL7Ih1CTN4Fd5YJ
8FcI9F6ZQYDcM1orQGu8t82SYdTqCThPAu6q4zR9NFgJQzoMbd3vLjBoQoHHcuWh
QGyctPYb4JoQnQ63y4kMNYQJmXNOyoqMjYoBLV5cl9UO3P8mVGBXpmdT9OzBbI9d
twjlsFTh6FWAK2PLR0NzHlXieMSA8FnUjUVpI1prK7eUQ9A9gh0bSUovVf5EJNa2
4QIDAQAB
-----END PUBLIC KEY-----`

	tmpFile := filepath.Join(t.TempDir(), "key.pem")
	err := os.WriteFile(tmpFile, []byte(pemData), 0600)
	require.NoError(t, err)

	cfg, err := NewClientConfig(WithRSAEncoder(tmpFile))
	require.NoError(t, err)
	require.NotNil(t, cfg)
	require.NotNil(t, cfg.rsaEncoder)

	encrypted, err := cfg.rsaEncoder([]byte("hello"))
	require.NoError(t, err)
	require.NotEmpty(t, encrypted)
}

func TestWithClient_Http(t *testing.T) {
	cfg, err := NewClientConfig(WithClient("http://localhost"))
	require.NoError(t, err)
	require.NotNil(t, cfg.httpClient)
	require.Nil(t, cfg.grpcClient)
}

func TestWithClient_InvalidURL(t *testing.T) {
	_, err := NewClientConfig(WithClient("%%invalid"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid server URL")
}

func TestWithClient_UnsupportedScheme(t *testing.T) {
	_, err := NewClientConfig(WithClient("ftp://example.com"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "unsupported URL scheme")
}

func TestWithClient_Grpc_Success(t *testing.T) {
	cfg := &ClientConfig{}
	err := WithClient("grpc://localhost:12345")(cfg)
	require.NoError(t, err)
	require.NotNil(t, cfg.grpcClient)
}

func TestWithClient_InvalidScheme(t *testing.T) {
	cfg := &ClientConfig{}
	err := WithClient("ftp://localhost:12345")(cfg)
	require.Error(t, err)
}
