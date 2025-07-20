package grpc

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure" // << import this
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/metadata"
)

// New creates a gRPC client with optional configuration
func New(target string, options ...Opt) (*grpc.ClientConn, error) {
	var dialOpts []grpc.DialOption

	var err error
	for _, opt := range options {
		dialOpts, err = opt(dialOpts)
		if err != nil {
			return nil, err
		}
	}

	return grpc.Dial(target, dialOpts...) // changed grpc.NewClient -> grpc.Dial
}

// Opt modifies grpc.DialOptions
type Opt func([]grpc.DialOption) ([]grpc.DialOption, error)

// WithInsecure disables transport security (useful for local or test)
func WithInsecure() Opt {
	return func(opts []grpc.DialOption) ([]grpc.DialOption, error) {
		return append(opts, grpc.WithTransportCredentials(insecure.NewCredentials())), nil
	}
}

// WithGRPCKeepaliveParams sets keepalive parameters
func WithKeepaliveParams(params keepalive.ClientParameters) Opt {
	return func(opts []grpc.DialOption) ([]grpc.DialOption, error) {
		return append(opts, grpc.WithKeepaliveParams(params)), nil
	}
}

// WithGRPCTransportCredentials sets custom transport credentials
func WithTransportCredentials(creds credentials.TransportCredentials) Opt {
	return func(opts []grpc.DialOption) ([]grpc.DialOption, error) {
		return append(opts, grpc.WithTransportCredentials(creds)), nil
	}
}

// WithTLSClientCert loads and sets a TLS cert/key pair
func WithTLSClientCert(certFile, keyFile string) Opt {
	return func(opts []grpc.DialOption) ([]grpc.DialOption, error) {
		cert, err := tls.LoadX509KeyPair(certFile, keyFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load client cert/key: %w", err)
		}

		rootCAs, err := x509.SystemCertPool()
		if err != nil {
			rootCAs = x509.NewCertPool()
		}

		tlsConfig := &tls.Config{
			Certificates: []tls.Certificate{cert},
			RootCAs:      rootCAs,
		}

		creds := credentials.NewTLS(tlsConfig)
		return append(opts, grpc.WithTransportCredentials(creds)), nil
	}
}

// UnaryInterceptorWithToken returns a unary client interceptor that injects
// the "authorization" metadata with the Bearer token.
func WithUnaryInterceptorToken(token string) grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply any,
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		md := metadata.Pairs("authorization", "Bearer "+token)
		ctx = metadata.NewOutgoingContext(ctx, md)
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

// StreamInterceptorWithToken returns a stream client interceptor that injects
// the "authorization" metadata with the Bearer token.
func WithStreamInterceptorToken(token string) grpc.StreamClientInterceptor {
	return func(
		ctx context.Context,
		desc *grpc.StreamDesc,
		cc *grpc.ClientConn,
		method string,
		streamer grpc.Streamer,
		opts ...grpc.CallOption,
	) (grpc.ClientStream, error) {
		md := metadata.Pairs("authorization", "Bearer "+token)
		ctx = metadata.NewOutgoingContext(ctx, md)
		return streamer(ctx, desc, cc, method, opts...)
	}
}
