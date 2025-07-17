package config

import (
	"errors"

	"github.com/sbilibin2017/gophkeeper/internal/configs"
	"github.com/sbilibin2017/gophkeeper/internal/configs/clients"
	"github.com/sbilibin2017/gophkeeper/internal/configs/scheme"
)

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
