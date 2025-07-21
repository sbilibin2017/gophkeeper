package grpc

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
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

	conn, err := grpc.NewClient(target, dialOpts...)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

// TLSCert represents a certificate and key file pair.
type TLSCert struct {
	CertFile string
	KeyFile  string
}

// WithTLSCert returns a DialOption for TLS credentials based on TLSCert.
func WithTLSCert(opts ...TLSCert) Opt {
	return func() (grpc.DialOption, error) {
		for _, certPair := range opts {
			if certPair.CertFile != "" && certPair.KeyFile != "" {
				cert, err := tls.LoadX509KeyPair(certPair.CertFile, certPair.KeyFile)
				if err != nil {
					return nil, fmt.Errorf("failed to load TLS cert/key: %w", err)
				}

				certPool, err := x509.SystemCertPool()
				if err != nil {
					return nil, fmt.Errorf("failed to load system cert pool: %w", err)
				}

				tlsConfig := &tls.Config{
					Certificates: []tls.Certificate{cert},
					RootCAs:      certPool,
				}

				creds := credentials.NewTLS(tlsConfig)
				return grpc.WithTransportCredentials(creds), nil
			}
		}
		return nil, nil
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

// WithToken adds a bearer token as per-RPC credentials.
func WithToken(opts ...string) Opt {
	return func() (grpc.DialOption, error) {
		for _, token := range opts {
			if token != "" {
				return grpc.WithPerRPCCredentials(tokenAuth{token: token}), nil
			}
		}
		return nil, nil
	}
}

// RetryPolicy defines gRPC retry configuration.
type RetryPolicy struct {
	Count   int
	Wait    time.Duration
	MaxWait time.Duration
}

// WithRetryPolicy sets retry configuration using the first valid RetryPolicy.
func WithRetryPolicy(opts ...RetryPolicy) Opt {
	return func() (grpc.DialOption, error) {
		for _, rp := range opts {
			if rp.Count > 0 || rp.Wait > 0 || rp.MaxWait > 0 {
				count := rp.Count
				wait := rp.Wait
				maxWait := rp.MaxWait

				// Convert to gRPC JSON duration format
				initialBackoff := fmt.Sprintf("%.3fs", wait.Seconds())
				maxBackoff := fmt.Sprintf("%.3fs", maxWait.Seconds())

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
				}`, count, initialBackoff, maxBackoff)

				return grpc.WithDefaultServiceConfig(cfg), nil
			}
		}
		return nil, nil
	}
}
