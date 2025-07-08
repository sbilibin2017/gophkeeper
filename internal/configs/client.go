package configs

import (
	"fmt"
	"path/filepath"
	"runtime"

	"github.com/go-resty/resty/v2"
	"github.com/jmoiron/sqlx"

	"github.com/sbilibin2017/gophkeeper/internal/configs/client"
	"github.com/sbilibin2017/gophkeeper/internal/configs/db"

	"google.golang.org/grpc"

	_ "modernc.org/sqlite"
)

// ClientConfig holds configuration for HTTP client, gRPC client, and database connection.
type ClientConfig struct {
	HTTPClient *resty.Client
	GRPCClient *grpc.ClientConn
	DB         *sqlx.DB
}

// ClientConfigOpt defines a function type for configuring ClientConfig.
type ClientConfigOpt func(*ClientConfig) error

// NewClientConfig creates a new ClientConfig by applying given options.
// Returns an error if any option returns an error.
func NewClientConfig(opts ...ClientConfigOpt) (*ClientConfig, error) {
	c := &ClientConfig{}
	for _, opt := range opts {
		if err := opt(c); err != nil {
			return nil, err
		}
	}
	return c, nil
}

// WithDB configures ClientConfig to use an SQLite database
// located in a file named "db.sqlite" inside the current file's directory.
func WithDB() ClientConfigOpt {
	return func(c *ClientConfig) error {
		_, currentFile, _, ok := runtime.Caller(0)
		if !ok {
			return fmt.Errorf("cannot get current file path")
		}
		dir := filepath.Dir(currentFile)

		databaseDSN := filepath.Join(dir, "db.sqlite")

		dbInstance, err := db.NewDB(databaseDSN)
		if err != nil {
			return err
		}
		c.DB = dbInstance
		return nil
	}
}

// WithHTTPClient configures ClientConfig to use an HTTP client
// with the base URL specified by serverURL.
func WithHTTPClient(serverURL string) ClientConfigOpt {
	return func(c *ClientConfig) error {
		client, err := client.NewHTTPClient(serverURL)
		if err != nil {
			return err
		}
		c.HTTPClient = client
		return nil
	}
}

// WithGRPCClient configures ClientConfig to use a gRPC client.
func WithGRPCClient(serverURL string) ClientConfigOpt {
	return func(c *ClientConfig) error {
		grpcClient, err := client.NewGRPCClient(serverURL)
		if err != nil {
			return err
		}
		c.GRPCClient = grpcClient
		return nil
	}
}
