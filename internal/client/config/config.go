package config

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/jmoiron/sqlx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	_ "modernc.org/sqlite"
)

// Config holds application-wide clients.
type Config struct {
	DB         *sqlx.DB
	HTTPClient *resty.Client
	GRPCClient *grpc.ClientConn
}

// Opt defines a functional option for Config initialization.
type Opt func(*Config) error

// NewConfig initializes a Config with optional components.
func NewConfig(opts ...Opt) (*Config, error) {
	cfg := &Config{}

	for _, opt := range opts {
		if err := opt(cfg); err != nil {
			return nil, err
		}
	}

	return cfg, nil
}

// WithDB connects to SQLite with optional custom DSN.
func WithDB() Opt {
	return func(cfg *Config) error {
		db, err := sqlx.Connect("sqlite", "client.db")
		if err != nil {
			return err
		}

		cfg.DB = db
		return nil
	}
}

func WithHTTPClient(baseURL, certPath, keyPath, token string) Opt {
	return func(cfg *Config) error {
		client := resty.New().
			SetBaseURL(baseURL)

		// Load and apply TLS certificates if provided
		if certPath != "" && keyPath != "" {
			cert, err := tls.LoadX509KeyPair(certPath, keyPath)
			if err != nil {
				return err
			}

			tlsConfig := &tls.Config{
				Certificates: []tls.Certificate{cert},
			}

			client.SetTransport(&http.Transport{
				TLSClientConfig: tlsConfig,
			})
		}

		// Set Authorization token if provided
		if token != "" {
			client.OnBeforeRequest(func(c *resty.Client, r *resty.Request) error {
				r.SetHeader("Authorization", "Bearer "+token)
				return nil
			})
		}

		// Configure retries: example values, adjust as needed
		client.
			SetRetryCount(3).
			SetRetryWaitTime(1 * time.Second).
			SetRetryMaxWaitTime(5 * time.Second).
			AddRetryCondition(func(r *resty.Response, err error) bool {
				return err != nil || (r.StatusCode() >= 500 && r.StatusCode() < 600)
			})

		cfg.HTTPClient = client
		return nil
	}
}

func WithGRPCClient(baseURL, certPath, keyPath, token string) Opt {
	return func(cfg *Config) error {
		var opts []grpc.DialOption

		// Setup transport credentials: TLS or insecure
		if certPath != "" && keyPath != "" {
			cert, err := tls.LoadX509KeyPair(certPath, keyPath)
			if err != nil {
				return err
			}
			tlsConfig := &tls.Config{Certificates: []tls.Certificate{cert}}
			creds := credentials.NewTLS(tlsConfig)
			opts = append(opts, grpc.WithTransportCredentials(creds))
		} else {
			opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
		}

		// Retry interceptor (3 retries, exponential backoff)
		retryUnary := retryInterceptor(3, 100*time.Millisecond)

		// Auth token interceptor if token provided
		var unaryInterceptor grpc.UnaryClientInterceptor
		if token != "" {
			tokenUnary := func(
				ctx context.Context,
				method string,
				req, reply any,
				cc *grpc.ClientConn,
				invoker grpc.UnaryInvoker,
				callOpts ...grpc.CallOption,
			) error {
				ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+token)
				return invoker(ctx, method, req, reply, cc, callOpts...)
			}

			unaryInterceptor = func(
				ctx context.Context,
				method string,
				req, reply any,
				cc *grpc.ClientConn,
				invoker grpc.UnaryInvoker,
				callOpts ...grpc.CallOption,
			) error {
				return retryUnary(ctx, method, req, reply, cc,
					func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
						return tokenUnary(ctx, method, req, reply, cc, invoker, opts...)
					}, callOpts...)
			}

			streamInterceptor := func(
				ctx context.Context,
				desc *grpc.StreamDesc,
				cc *grpc.ClientConn,
				method string,
				streamer grpc.Streamer,
				callOpts ...grpc.CallOption,
			) (grpc.ClientStream, error) {
				ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+token)
				return streamer(ctx, desc, cc, method, callOpts...)
			}

			opts = append(opts,
				grpc.WithUnaryInterceptor(unaryInterceptor),
				grpc.WithStreamInterceptor(streamInterceptor),
			)
		} else {
			opts = append(opts, grpc.WithUnaryInterceptor(retryUnary))
		}

		// Custom dialer with context
		opts = append(opts, grpc.WithContextDialer(func(ctx context.Context, addr string) (net.Conn, error) {
			return (&net.Dialer{}).DialContext(ctx, "tcp", addr)
		}))

		// Create the client connection
		conn, err := grpc.NewClient(baseURL, opts...)
		if err != nil {
			return err
		}

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		// Trigger connection attempts
		conn.Connect()

		// Wait for Ready state or timeout
		for {
			state := conn.GetState()
			if state == connectivity.Ready {
				break
			}
			if !conn.WaitForStateChange(ctx, state) {
				conn.Close()
				return fmt.Errorf("grpc connection failed to become ready")
			}
		}

		if cfg.GRPCClient != nil {
			_ = cfg.GRPCClient.Close()
		}
		cfg.GRPCClient = conn
		return nil
	}
}

func retryInterceptor(maxRetries int, baseBackoff time.Duration) grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply any,
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		var err error
		backoff := baseBackoff
		for i := 0; i < maxRetries; i++ {
			err = invoker(ctx, method, req, reply, cc, opts...)
			if err == nil {
				return nil
			}
			time.Sleep(backoff)
			backoff *= 2
		}
		return err
	}
}
