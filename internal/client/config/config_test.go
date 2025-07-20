package config

import (
	"errors"
	"os"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"

	grpcClient "github.com/sbilibin2017/gophkeeper/internal/configs/clients/grpc"
	httpClient "github.com/sbilibin2017/gophkeeper/internal/configs/clients/http"
	"github.com/sbilibin2017/gophkeeper/internal/configs/db"
)

func cleanupDBFile(t *testing.T) {
	t.Helper()
	t.Cleanup(func() {
		err := os.Remove("client.db")
		if err != nil && !os.IsNotExist(err) {
			t.Errorf("failed to remove client.db: %v", err)
		}
	})
}

func TestNewConfig_DBError(t *testing.T) {
	origDBNew := dbNew
	defer func() { dbNew = origDBNew }()
	dbNew = func(driver, dsn string, opts ...db.Opt) (*sqlx.DB, error) {
		return nil, errors.New("db open error")
	}

	cfg, err := NewConfig()
	assert.Nil(t, cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "db open error")
}

func TestNewConfig_DefaultSuccess(t *testing.T) {
	cleanupDBFile(t)
	cfg, err := NewConfig()
	require.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.NotNil(t, cfg.DB)
	assert.Nil(t, cfg.HTTPClient)
	assert.Nil(t, cfg.GRPCClient)
}

func TestWithHTTPClient_Success(t *testing.T) {
	mockURL := "https://example.com"

	origHTTPNew := httpClientNew
	defer func() { httpClientNew = origHTTPNew }()

	httpClientNew = func(url string, opts ...httpClient.Opt) (*resty.Client, error) {
		assert.Equal(t, mockURL, url)
		return resty.New(), nil
	}

	cfg := &Config{}
	err := WithHTTPClient(mockURL, "", "", "")(cfg)
	require.NoError(t, err)
	assert.NotNil(t, cfg.HTTPClient)
	assert.Nil(t, cfg.GRPCClient)
}

func TestWithHTTPClient_WithTLSAndToken(t *testing.T) {
	mockURL := "https://example.com"
	cert := "certdata"
	key := "keydata"
	token := "tok"

	origHTTPNew := httpClientNew
	defer func() { httpClientNew = origHTTPNew }()

	httpClientNew = func(url string, opts ...httpClient.Opt) (*resty.Client, error) {
		assert.Equal(t, mockURL, url)
		require.Len(t, opts, 2)
		// Can't check internals easily but expect two options (TLS and Token)
		return resty.New(), nil
	}

	cfg := &Config{}
	err := WithHTTPClient(mockURL, cert, key, token)(cfg)
	require.NoError(t, err)
	assert.NotNil(t, cfg.HTTPClient)
}

func TestWithHTTPClient_ErrorFromNew(t *testing.T) {
	mockURL := "https://example.com"

	origHTTPNew := httpClientNew
	defer func() { httpClientNew = origHTTPNew }()

	httpClientNew = func(url string, opts ...httpClient.Opt) (*resty.Client, error) {
		return nil, errors.New("http client error")
	}

	cfg := &Config{}
	err := WithHTTPClient(mockURL, "", "", "")(cfg)
	assert.Error(t, err)
	assert.Nil(t, cfg.HTTPClient)
}

func TestWithGRPCClient_InsecureNoToken(t *testing.T) {
	mockURL := "grpc://localhost:1234"

	origGRPCNew := grpcClientNew
	defer func() { grpcClientNew = origGRPCNew }()

	grpcClientNew = func(addr string, opts ...grpcClient.Opt) (*grpc.ClientConn, error) {
		assert.Equal(t, "localhost:1234", addr)
		require.Len(t, opts, 1) // WithInsecure only
		return &grpc.ClientConn{}, nil
	}

	cfg := &Config{}
	err := WithGRPCClient(mockURL, "", "", "")(cfg)
	require.NoError(t, err)
	assert.NotNil(t, cfg.GRPCClient)
	assert.Nil(t, cfg.HTTPClient)
}

func TestWithGRPCClient_TLSWithToken(t *testing.T) {
	mockURL := "grpc://localhost:1234"
	cert := "certdata"
	key := "keydata"
	token := "tok"

	origGRPCNew := grpcClientNew
	defer func() { grpcClientNew = origGRPCNew }()

	grpcClientNew = func(addr string, opts ...grpcClient.Opt) (*grpc.ClientConn, error) {
		assert.Equal(t, "localhost:1234", addr)
		// TLS cert + unary interceptor + stream interceptor, so at least 3 opts
		require.GreaterOrEqual(t, len(opts), 3)
		return &grpc.ClientConn{}, nil
	}

	cfg := &Config{}
	err := WithGRPCClient(mockURL, cert, key, token)(cfg)
	require.NoError(t, err)
	assert.NotNil(t, cfg.GRPCClient)
}

func TestWithGRPCClient_ErrorFromNew(t *testing.T) {
	mockURL := "grpc://localhost:1234"

	origGRPCNew := grpcClientNew
	defer func() { grpcClientNew = origGRPCNew }()

	grpcClientNew = func(addr string, opts ...grpcClient.Opt) (*grpc.ClientConn, error) {
		return nil, errors.New("grpc client error")
	}

	cfg := &Config{}
	err := WithGRPCClient(mockURL, "", "", "")(cfg)
	assert.Error(t, err)
	assert.Nil(t, cfg.GRPCClient)
}

func TestStripScheme(t *testing.T) {
	host, err := stripScheme("grpc://localhost:1234")
	require.NoError(t, err)
	assert.Equal(t, "localhost:1234", host)

	host, err = stripScheme("https://example.com")
	require.NoError(t, err)
	assert.Equal(t, "example.com", host)

	_, err = stripScheme("://invalid-url")
	assert.Error(t, err)
}
