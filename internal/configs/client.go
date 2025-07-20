package configs

import (
	"net/url"

	"github.com/go-resty/resty/v2"
	"github.com/jmoiron/sqlx"
	"google.golang.org/grpc"

	"github.com/sbilibin2017/gophkeeper/internal/configs/db"

	grpcClient "github.com/sbilibin2017/gophkeeper/internal/configs/clients/grpc"
	"github.com/sbilibin2017/gophkeeper/internal/configs/clients/http"
	httpClient "github.com/sbilibin2017/gophkeeper/internal/configs/clients/http"
)

const (
	// clientDriverName is the default database driver name used by the client.
	clientDriverName = "sqlite"
	// clientDSN is the default data source name for the client database.
	clientDSN = "client.db"
)

// package-level variables to allow overriding in tests
var (
	httpClientNew = httpClient.New
	grpcClientNew = grpcClient.New
	dbNew         = db.NewDB
)

// ClientConfig holds configuration for the client, including database and network clients.
type ClientConfig struct {
	DB         *sqlx.DB
	HTTPClient *resty.Client
	GRPCClient *grpc.ClientConn
}

// ClientConfigOpt defines a functional option for configuring ClientConfig.
type ClientConfigOpt func(*ClientConfig) error

// NewClientConfig creates a new ClientConfig with default settings, applying any
// provided functional options to customize the configuration.
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

// WithHTTPClient returns a ClientConfigOpt that configures the HTTP client with
// the provided authURL, optional TLS client certificate/key, and bearer token.
func WithHTTPClient(authURL, tlsClientCert, tlsClientKey, token string) ClientConfigOpt {
	return func(cfg *ClientConfig) error {
		httpOpts := []http.Opt{}
		if tlsClientCert != "" && tlsClientKey != "" {
			httpOpts = append(httpOpts, http.WithTLSClientCert(tlsClientCert, tlsClientKey))
		}
		if token != "" {
			httpOpts = append(httpOpts, http.WithToken(token))
		}

		client, err := httpClientNew(authURL, httpOpts...)
		if err != nil {
			return err
		}

		cfg.HTTPClient = client
		return nil
	}
}

// WithGRPCClient returns a ClientConfigOpt that configures the gRPC client with
// the given authURL, optional TLS client certificate/key, and bearer token.
// It handles creating secure or insecure connections and applies token interceptors.
func WithGRPCClient(authURL, tlsClientCert, tlsClientKey, token string) ClientConfigOpt {
	return func(cfg *ClientConfig) error {
		grpcOpts := []grpcClient.Opt{}

		if tlsClientCert != "" && tlsClientKey != "" {
			grpcOpts = append(grpcOpts, grpcClient.WithTLSClientCert(tlsClientCert, tlsClientKey))
		} else {
			grpcOpts = append(grpcOpts, grpcClient.WithInsecure())
		}

		if token != "" {
			grpcOpts = append(grpcOpts,
				grpcClient.WithUnaryInterceptor(grpcClient.WithUnaryInterceptorToken(token)),
				grpcClient.WithStreamInterceptor(grpcClient.WithStreamInterceptorToken(token)),
			)
		}

		addr, err := stripScheme(authURL)
		if err != nil {
			return err
		}

		conn, err := grpcClientNew(addr, grpcOpts...)
		if err != nil {
			return err
		}

		cfg.GRPCClient = conn
		return nil
	}
}

// stripScheme parses the raw URL string and returns only the host portion (host:port),
// removing the scheme (e.g., "http://", "grpc://").
// Returns an error if the URL cannot be parsed.
func stripScheme(rawURL string) (string, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	return parsed.Host, nil
}
