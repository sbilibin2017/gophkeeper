package configs

import (
	"github.com/go-resty/resty/v2"
	"github.com/jmoiron/sqlx"
	"github.com/sbilibin2017/gophkeeper/internal/configs/clients"
	"github.com/sbilibin2017/gophkeeper/internal/configs/db"
	"google.golang.org/grpc"
)

// ClientConfig holds client configuration including HTTP client, gRPC client, and database connection.
type ClientConfig struct {
	HTTPClient *resty.Client
	GRPCClient *grpc.ClientConn
	DB         *sqlx.DB
}

// ClientConfigOpt defines a function type for applying options to ClientConfig.
type ClientConfigOpt func(*ClientConfig) error

// NewClientConfig creates a new ClientConfig and applies the provided options.
func NewClientConfig(opts ...ClientConfigOpt) (*ClientConfig, error) {
	cfg := &ClientConfig{}
	for _, opt := range opts {
		if err := opt(cfg); err != nil {
			return nil, err
		}
	}
	return cfg, nil
}

// WithClientConfigHTTPClient creates an HTTP client with the first non-empty baseURL and optional HTTP client options.
func WithClientConfigHTTPClient(baseURL string, httpOpts ...clients.HTTPClientOption) ClientConfigOpt {
	return func(cfg *ClientConfig) error {
		if baseURL == "" {
			return nil
		}
		client, err := clients.NewHTTPClient(baseURL, httpOpts...)
		if err != nil {
			return err
		}
		cfg.HTTPClient = client
		return nil
	}
}

// WithClientConfigGRPCClient creates a gRPC client with the given non-empty address and optional gRPC client options.
func WithClientConfigGRPCClient(addr string, grpcOpts ...clients.GRPCClientOption) ClientConfigOpt {
	return func(cfg *ClientConfig) error {
		if addr == "" {
			return nil
		}
		conn, err := clients.NewGRPCClient(addr, grpcOpts...)
		if err != nil {
			return err
		}
		cfg.GRPCClient = conn
		return nil
	}
}

// WithClientConfigDB connects to a database using the specified driver and DSN.
func WithClientConfigDB() ClientConfigOpt {
	return func(cfg *ClientConfig) error {
		conn, err := db.NewDB("sqlite", "client.db")
		if err != nil {
			return err
		}
		cfg.DB = conn
		return nil
	}
}

// WithClientConfigDBWithMigrations connects to a SQLite database with the provided DSN and applies migrations.
func WithClientConfigDBWithMigrations(dsn string) ClientConfigOpt {
	return func(cfg *ClientConfig) error {
		if dsn == "" {
			return nil
		}
		conn, err := db.NewDB("sqlite", dsn)
		if err != nil {
			return err
		}
		cfg.DB = conn
		return nil
	}
}
