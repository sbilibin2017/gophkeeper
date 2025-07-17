package clients

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

// GRPCClientOption defines a function type to configure grpc.DialOption.
type GRPCClientOption func(*grpcDialOptions) error

type grpcDialOptions struct {
	dialOpts []grpc.DialOption
}

// WithGRPCKeepaliveParams sets the keepalive parameters for the gRPC client.
// Parameters:
// - params: keepalive.ClientParameters struct specifying the keepalive settings.
// Returns:
// - a GRPCClientOption function to apply the keepalive parameters.
func WithGRPCKeepaliveParams(params keepalive.ClientParameters) GRPCClientOption {
	return func(opts *grpcDialOptions) error {
		opts.dialOpts = append(opts.dialOpts, grpc.WithKeepaliveParams(params))
		return nil
	}
}

// WithGRPCTransportCredentials sets the transport credentials for the gRPC client.
// Parameters:
// - creds: transport credentials (e.g., TLS credentials).
// Returns:
// - a GRPCClientOption function to apply the transport credentials.
func WithGRPCTransportCredentials(creds credentials.TransportCredentials) GRPCClientOption {
	return func(opts *grpcDialOptions) error {
		opts.dialOpts = append(opts.dialOpts, grpc.WithTransportCredentials(creds))
		return nil
	}
}

// WithGRPCTLSClientCert loads a client TLS certificate and private key, then configures the client with them.
// Parameters:
// - certFile: path to the client certificate file.
// - keyFile: path to the client private key file.
// Returns:
// - a GRPCClientOption function that applies the TLS client cert.
// - an error if loading the certificate or key fails.
func WithGRPCTLSClientCert(certFile, keyFile string) GRPCClientOption {
	return func(opts *grpcDialOptions) error {
		certificate, err := tls.LoadX509KeyPair(certFile, keyFile)
		if err != nil {
			return fmt.Errorf("failed to load client certificate/key: %w", err)
		}

		// Load system root CA pool
		rootCAs, err := x509.SystemCertPool()
		if err != nil {
			rootCAs = x509.NewCertPool()
		}

		// Create TLS config with client cert and root CAs
		tlsConfig := &tls.Config{
			Certificates: []tls.Certificate{certificate},
			RootCAs:      rootCAs,
		}

		creds := credentials.NewTLS(tlsConfig)
		opts.dialOpts = append(opts.dialOpts, grpc.WithTransportCredentials(creds))
		return nil
	}
}

// NewGRPCClient creates a new gRPC client connection to the specified target with optional configurations.
// Default options include:
// - insecure transport credentials (if not overridden)
// - keepalive parameters set to 10s interval, 3s timeout, and permit without stream enabled.
// Parameters:
// - target: the server address to connect to.
// - opts: optional GRPCClientOption functions to customize the client connection.
// Returns:
// - a grpc.ClientConn instance.
// - an error if the connection fails.
func NewGRPCClient(target string, opts ...GRPCClientOption) (*grpc.ClientConn, error) {
	options := &grpcDialOptions{}

	// Set default dial options
	options.dialOpts = append(options.dialOpts,
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                10 * time.Second,
			Timeout:             3 * time.Second,
			PermitWithoutStream: true,
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	// Apply user options that can override defaults
	for _, opt := range opts {
		if err := opt(options); err != nil {
			return nil, err
		}
	}

	conn, err := grpc.NewClient(target, options.dialOpts...)
	if err != nil {
		return nil, err
	}

	return conn, nil
}
