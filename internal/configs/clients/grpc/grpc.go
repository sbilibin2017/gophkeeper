package grpc

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// Opt defines a function type returning a grpc.DialOption and possibly an error.
type Opt func() (grpc.DialOption, error)

// New creates a gRPC ClientConn applying all options as DialOptions.
func New(target string, opts ...Opt) (*grpc.ClientConn, error) {
	var dialOpts []grpc.DialOption

	for _, opt := range opts {
		dialOpt, err := opt()
		if err != nil {
			return nil, err
		}
		if dialOpt != nil {
			dialOpts = append(dialOpts, dialOpt)
		}
	}

	conn, err := grpc.Dial(target, dialOpts...)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

// WithTLSCert returns a DialOption for TLS credentials using one or more root CA cert files.
func WithTLSCert(certFiles ...string) Opt {
	return func() (grpc.DialOption, error) {
		if len(certFiles) == 0 {
			return nil, nil
		}

		certPool := x509.NewCertPool()
		for _, certFile := range certFiles {
			if certFile == "" {
				continue
			}

			certPEM, err := os.ReadFile(certFile)
			if err != nil {
				return nil, fmt.Errorf("failed to read cert file %q: %w", certFile, err)
			}

			if ok := certPool.AppendCertsFromPEM(certPEM); !ok {
				return nil, fmt.Errorf("failed to append cert to pool from file %q", certFile)
			}
		}

		tlsConfig := &tls.Config{
			RootCAs: certPool,
		}

		creds := credentials.NewTLS(tlsConfig)
		return grpc.WithTransportCredentials(creds), nil
	}
}

// tokenAuth implements grpc.PerRPCCredentials to inject a bearer token.
type tokenAuth struct {
	token string
}

func (t tokenAuth) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	return map[string]string{
		"authorization": "Bearer " + t.token,
	}, nil
}

func (t tokenAuth) RequireTransportSecurity() bool {
	return true
}

// WithAuthToken adds a bearer token as per-RPC credentials.
func WithAuthToken(token string) Opt {
	return func() (grpc.DialOption, error) {
		if token == "" {
			return nil, nil
		}
		return grpc.WithPerRPCCredentials(tokenAuth{token: token}), nil
	}
}

// RetryPolicy defines gRPC retry configuration.
type RetryPolicy struct {
	Count   int
	Wait    time.Duration
	MaxWait time.Duration
}

// WithRetryPolicy sets retry configuration using the first valid RetryPolicy.
func WithRetryPolicy(rp RetryPolicy) Opt {
	return func() (grpc.DialOption, error) {
		if rp.Count <= 0 && rp.Wait <= 0 && rp.MaxWait <= 0 {
			return nil, nil
		}

		initialBackoff := fmt.Sprintf("%.3fs", rp.Wait.Seconds())
		maxBackoff := fmt.Sprintf("%.3fs", rp.MaxWait.Seconds())

		cfg := fmt.Sprintf(`{
			"methodConfig": [{
				"name": [{"service": ".*"}],
				"retryPolicy": {
					"maxAttempts": %d,
					"initialBackoff": "%s",
					"maxBackoff": "%s",
					"backoffMultiplier": 2,
					"retryableStatusCodes": ["UNAVAILABLE"]
				}
			}]
		}`, rp.Count, initialBackoff, maxBackoff)

		return grpc.WithDefaultServiceConfig(cfg), nil
	}
}
