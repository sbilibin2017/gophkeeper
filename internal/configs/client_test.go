package configs

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

var (
	mockHTTPURL = "https://localhost:1234"
	mockGRPCURL = "grpc://localhost:50051"
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

func TestNewClientConfig_DBNewError(t *testing.T) {
	// Save original function and restore after test
	origDBNew := dbNew
	defer func() { dbNew = origDBNew }()

	// Override dbNew to simulate error
	dbNew = func(driver string, dsn string, opts ...db.Opt) (*sqlx.DB, error) {
		return nil, errors.New("failed to open DB")
	}

	cfg, err := NewClientConfig()
	assert.Nil(t, cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to open DB")
}

func TestNewClientConfig_Default(t *testing.T) {
	cleanupDBFile(t)
	cfg, err := NewClientConfig()
	require.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.NotNil(t, cfg.DB)
	assert.Nil(t, cfg.HTTPClient)
	assert.Nil(t, cfg.GRPCClient)
}

func TestNewClientConfig_WithCustomDB(t *testing.T) {
	cleanupDBFile(t)
	db := &sqlx.DB{} // fake/mock, just a pointer to satisfy type
	cfg, err := NewClientConfig(WithCustomDB(db))
	require.NoError(t, err)
	assert.Equal(t, db, cfg.DB)
}

func TestNewClientConfig_WithCustomHTTPClient(t *testing.T) {
	cleanupDBFile(t)
	client := resty.New()
	cfg, err := NewClientConfig(WithCustomHTTPClient(client))
	require.NoError(t, err)
	assert.Equal(t, client, cfg.HTTPClient)
}

func TestNewClientConfig_WithCustomGRPCClient(t *testing.T) {
	cleanupDBFile(t)
	var conn *grpc.ClientConn = nil // mock, nil is fine here
	cfg, err := NewClientConfig(WithCustomGRPCClient(conn))
	require.NoError(t, err)
	assert.Equal(t, conn, cfg.GRPCClient)
}

func TestNewClientConfig_WithAuthURL_HTTP(t *testing.T) {
	cleanupDBFile(t)
	cfg, err := NewClientConfig(WithAuthURL(mockHTTPURL, "", ""))
	require.NoError(t, err)
	assert.NotNil(t, cfg.HTTPClient)
	assert.Nil(t, cfg.GRPCClient)
}

func TestNewClientConfig_WithAuthURL_GRPC_Insecure(t *testing.T) {
	cleanupDBFile(t)
	cfg, err := NewClientConfig(WithAuthURL(mockGRPCURL, "", ""))
	require.NoError(t, err)
	assert.NotNil(t, cfg.GRPCClient)
	assert.Nil(t, cfg.HTTPClient)
}

func TestNewClientConfig_UnsupportedScheme(t *testing.T) {
	cleanupDBFile(t)
	cfg, err := NewClientConfig(WithAuthURL("ftp://server", "", ""))
	require.Error(t, err)
	assert.Nil(t, cfg)
}

func TestWithAuthURL_HTTPClientError(t *testing.T) {
	// Save original and defer restore
	origHTTPNew := httpClientNew
	defer func() { httpClientNew = origHTTPNew }()

	// Override httpClientNew to always return error
	httpClientNew = func(url string, opts ...httpClient.Opt) (*resty.Client, error) {
		return nil, errors.New("http client creation failed")
	}

	cfg := &ClientConfig{}
	err := WithAuthURL("https://example.com", "", "")(cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "http client creation failed")
	assert.Nil(t, cfg.HTTPClient)
}

func TestWithAuthURL_GRPCClientError(t *testing.T) {
	// Save original and defer restore
	origGRPCNew := grpcClientNew
	defer func() { grpcClientNew = origGRPCNew }()

	// Override grpcClientNew to always return error
	grpcClientNew = func(url string, opts ...grpcClient.Opt) (*grpc.ClientConn, error) {
		return nil, errors.New("grpc client creation failed")
	}

	cfg := &ClientConfig{}
	err := WithAuthURL("grpc://localhost:50051", "", "")(cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "grpc client creation failed")
	assert.Nil(t, cfg.GRPCClient)
}

func TestWithAuthURL_HTTP_NoTLS(t *testing.T) {
	// Mock httpClientNew to verify options passed and simulate client creation
	httpClientNew = func(url string, opts ...httpClient.Opt) (*resty.Client, error) {
		require.Equal(t, "https://example.com", url)
		require.Len(t, opts, 0) // No TLS options expected
		return resty.New(), nil
	}
	defer func() { httpClientNew = httpClient.New }()

	cfg := &ClientConfig{}
	err := WithAuthURL("https://example.com", "", "")(cfg)
	require.NoError(t, err)
	require.NotNil(t, cfg.HTTPClient)
	require.Nil(t, cfg.GRPCClient)
}

func TestWithAuthURL_HTTP_WithTLS(t *testing.T) {
	cert, key := "certdata", "keydata"

	var receivedOpts []httpClient.Opt

	httpClientNew = func(url string, opts ...httpClient.Opt) (*resty.Client, error) {
		require.Equal(t, "https://secure.com", url)
		receivedOpts = opts
		return resty.New(), nil
	}
	defer func() { httpClientNew = httpClient.New }()

	cfg := &ClientConfig{}
	err := WithAuthURL("https://secure.com", cert, key)(cfg)
	require.NoError(t, err)
	require.NotNil(t, cfg.HTTPClient)
	require.Nil(t, cfg.GRPCClient)

	// Check that TLS option is passed
	require.NotEmpty(t, receivedOpts)
}

func TestWithAuthURL_GRPC_NoTLS(t *testing.T) {
	grpcClientNew = func(url string, opts ...grpcClient.Opt) (*grpc.ClientConn, error) {
		require.Equal(t, "grpc://localhost:1234", url)
		require.Len(t, opts, 1) // Expect WithInsecure option only
		return &grpc.ClientConn{}, nil
	}
	defer func() { grpcClientNew = grpcClient.New }()

	cfg := &ClientConfig{}
	err := WithAuthURL("grpc://localhost:1234", "", "")(cfg)
	require.NoError(t, err)
	require.NotNil(t, cfg.GRPCClient)
	require.Nil(t, cfg.HTTPClient)
}

func TestWithAuthURL_GRPC_WithTLS(t *testing.T) {
	cert, key := "certdata", "keydata"

	var receivedOpts []grpcClient.Opt

	grpcClientNew = func(url string, opts ...grpcClient.Opt) (*grpc.ClientConn, error) {
		require.Equal(t, "grpc://secure-grpc", url)
		receivedOpts = opts
		return &grpc.ClientConn{}, nil
	}
	defer func() { grpcClientNew = grpcClient.New }()

	cfg := &ClientConfig{}
	err := WithAuthURL("grpc://secure-grpc", cert, key)(cfg)
	require.NoError(t, err)
	require.NotNil(t, cfg.GRPCClient)
	require.Nil(t, cfg.HTTPClient)

	// TLS option should be passed
	require.NotEmpty(t, receivedOpts)
}

func TestWithAuthURL_UnsupportedScheme(t *testing.T) {
	cfg := &ClientConfig{}
	err := WithAuthURL("ftp://unsupported", "", "")(cfg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "unsupported URL scheme")
	require.Nil(t, cfg.HTTPClient)
	require.Nil(t, cfg.GRPCClient)
}
