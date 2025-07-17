package config

import (
	"errors"

	"github.com/sbilibin2017/gophkeeper/internal/configs"
	"github.com/sbilibin2017/gophkeeper/internal/configs/clients"
	"github.com/sbilibin2017/gophkeeper/internal/configs/scheme"
)

// NewConfig creates a new ClientConfig based on the given authentication URL
// and optional TLS client certificate and key files.
//
// The function determines the scheme from the authURL (HTTP, HTTPS, or gRPC)
// and configures the client accordingly:
// - For HTTP/HTTPS, it configures an HTTP client with optional TLS certificates.
// - For gRPC, it configures a gRPC client with optional TLS certificates.
//
// Returns an error if the URL scheme is unsupported or configuration fails.
func NewConfig(
	authURL string,
	tlsClientCert string,
	tlsClientKey string,
) (*configs.ClientConfig, error) {
	var opts []configs.ClientConfigOpt

	opts = append(opts, configs.WithClientConfigDB())

	schm := scheme.GetSchemeFromURL(authURL)

	switch schm {
	case scheme.HTTP, scheme.HTTPS:
		httpOpts := []clients.HTTPClientOption{}
		if tlsClientCert != "" && tlsClientKey != "" {
			httpOpts = append(httpOpts, clients.WithHTTPTLSClientCert(tlsClientCert, tlsClientKey))
		}
		opts = append(opts, configs.WithClientConfigHTTPClient(authURL, httpOpts...))

	case scheme.GRPC:
		grpcOpts := []clients.GRPCClientOption{}
		if tlsClientCert != "" && tlsClientKey != "" {
			grpcOpts = append(grpcOpts, clients.WithGRPCTLSClientCert(tlsClientCert, tlsClientKey))
		}
		opts = append(opts, configs.WithClientConfigGRPCClient(authURL, grpcOpts...))

	default:
		return nil, errors.New("unsupported scheme: " + schm)
	}

	return configs.NewClientConfig(opts...)
}
