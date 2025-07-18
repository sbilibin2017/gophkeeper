package config

import (
	"errors"
	"fmt"

	"github.com/sbilibin2017/gophkeeper/internal/configs"
	"github.com/sbilibin2017/gophkeeper/internal/configs/clients"
	"github.com/sbilibin2017/gophkeeper/internal/configs/scheme"
)

// NewClientConfig creates a new client configuration based on the authURL,
// optional TLS client cert and key paths.
func NewClientConfig(authURL, tlsClientCert, tlsClientKey string) (*configs.ClientConfig, error) {
	var opts []configs.ClientConfigOpt

	// Always include DB config option (assuming you want local DB by default)
	opts = append(opts, configs.WithClientConfigDB())

	// If authURL is empty, just return config with DB only
	if authURL == "" {
		return configs.NewClientConfig(opts...)
	}

	// Detect scheme and add appropriate client config options
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

	cfg, err := configs.NewClientConfig(opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create client config: %w", err)
	}

	return cfg, nil
}
