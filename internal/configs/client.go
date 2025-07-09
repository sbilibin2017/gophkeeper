package configs

import (
	"fmt"
	"net/url"

	"github.com/go-resty/resty/v2"
	"github.com/jmoiron/sqlx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	_ "modernc.org/sqlite"
)

// ClientConfig holds configuration for HTTP client, gRPC client, and database connection.
type ClientConfig struct {
	HTTPClient *resty.Client
	GRPCClient *grpc.ClientConn
	DB         *sqlx.DB
	Token      string
}

// ClientConfigOpt defines a function type for configuring ClientConfig.
type ClientConfigOpt func(*ClientConfig) error

// NewClientConfig creates a new ClientConfig by applying given options.
// Returns an error if any option fails.
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
// located at the specified path.
func WithDB(pathToDB string) ClientConfigOpt {
	return func(c *ClientConfig) error {
		db, err := sqlx.Open("sqlite", pathToDB)
		if err != nil {
			return fmt.Errorf("failed to open database: %w", err)
		}

		if err := db.Ping(); err != nil {
			return fmt.Errorf("failed to connect to database: %w", err)
		}

		c.DB = db
		return nil
	}
}

// WithHTTPClient configures ClientConfig to use an HTTP client
// with the base URL specified by serverURL.
func WithHTTPClient(serverURL string) ClientConfigOpt {
	return func(c *ClientConfig) error {
		parsedURL, err := url.Parse(serverURL)
		if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
			return fmt.Errorf("invalid serverURL for HTTP client: %s", serverURL)
		}

		c.HTTPClient = resty.New().SetBaseURL(parsedURL.String())
		return nil
	}
}

// WithGRPCClient configures ClientConfig to use a gRPC client.
func WithGRPCClient(serverURL string) ClientConfigOpt {
	return func(c *ClientConfig) error {
		parsedURL, err := url.Parse(serverURL)
		if err != nil || parsedURL.Host == "" {
			return fmt.Errorf("invalid serverURL for gRPC client: %s", serverURL)
		}

		conn, err := grpc.NewClient(parsedURL.Host, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			return fmt.Errorf("failed to connect to gRPC server: %w", err)
		}

		c.GRPCClient = conn
		return nil
	}
}

// WithToken configures ClientConfig to use a JWT token.
func WithToken(token string) ClientConfigOpt {
	return func(c *ClientConfig) error {
		if token == "" {
			return nil
		}
		c.Token = token
		return nil
	}
}
