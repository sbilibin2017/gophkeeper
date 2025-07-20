package configs

import (
	"fmt"

	"github.com/go-resty/resty/v2"
	"github.com/jmoiron/sqlx"
	"google.golang.org/grpc"

	"github.com/sbilibin2017/gophkeeper/internal/configs/db"
	"github.com/sbilibin2017/gophkeeper/internal/configs/scheme"

	grpcClient "github.com/sbilibin2017/gophkeeper/internal/configs/clients/grpc"
	httpClient "github.com/sbilibin2017/gophkeeper/internal/configs/clients/http"
)

const (
	clientDriverName = "sqlite"
	clientDSN        = "client.db"
)

// package-level variables to allow overriding in tests
var (
	httpClientNew = httpClient.New
	grpcClientNew = grpcClient.New
	dbNew         = db.NewDB
)

// ClientConfig holds configuration for the client
type ClientConfig struct {
	DB         *sqlx.DB
	HTTPClient *resty.Client
	GRPCClient *grpc.ClientConn
}

// ClientConfigOpt defines a functional option for ClientConfig
type ClientConfigOpt func(*ClientConfig) error

// NewClientConfig creates a new ClientConfig and applies functional options
func NewClientConfig(opts ...ClientConfigOpt) (*ClientConfig, error) {
	cfg := &ClientConfig{}

	conn, err := dbNew(clientDriverName, clientDSN)
	if err != nil {
		return nil, err
	}
	cfg.DB = conn

	for _, opt := range opts {
		if err := opt(cfg); err != nil {
			return nil, err
		}
	}

	return cfg, nil
}

// WithCustomDB overrides the default DB connection
func WithCustomDB(conn *sqlx.DB) ClientConfigOpt {
	return func(cfg *ClientConfig) error {
		cfg.DB = conn
		return nil
	}
}

// WithCustomHTTPClient overrides the HTTP client
func WithCustomHTTPClient(client *resty.Client) ClientConfigOpt {
	return func(cfg *ClientConfig) error {
		cfg.HTTPClient = client
		return nil
	}
}

// WithCustomGRPCClient overrides the gRPC client
func WithCustomGRPCClient(conn *grpc.ClientConn) ClientConfigOpt {
	return func(cfg *ClientConfig) error {
		cfg.GRPCClient = conn
		return nil
	}
}

// WithAuthURL creates HTTP or gRPC clients based on URL scheme
func WithAuthURL(authURL, tlsClientCert, tlsClientKey string) ClientConfigOpt {
	return func(cfg *ClientConfig) error {
		schm := scheme.GetSchemeFromURL(authURL)

		switch schm {
		case scheme.HTTP, scheme.HTTPS:
			httpOpts := []httpClient.Opt{}
			if tlsClientCert != "" && tlsClientKey != "" {
				httpOpts = append(httpOpts, httpClient.WithTLSClientCert(tlsClientCert, tlsClientKey))
			}

			client, err := httpClientNew(authURL, httpOpts...)
			if err != nil {
				return err
			}
			cfg.HTTPClient = client

		case scheme.GRPC:
			grpcOpts := []grpcClient.Opt{}
			if tlsClientCert != "" && tlsClientKey != "" {
				grpcOpts = append(grpcOpts, grpcClient.WithTLSClientCert(tlsClientCert, tlsClientKey))
			} else {
				grpcOpts = append(grpcOpts, grpcClient.WithInsecure())
			}

			conn, err := grpcClientNew(authURL, grpcOpts...)
			if err != nil {
				return err
			}
			cfg.GRPCClient = conn

		default:
			return fmt.Errorf("unsupported URL scheme: %s", schm)
		}

		return nil
	}
}
