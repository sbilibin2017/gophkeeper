package main

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
	driverName = "sqlite"
	dsn        = "client.db"
)

type config struct {
	DB         *sqlx.DB
	HTTPClient *resty.Client
	GRPCClient *grpc.ClientConn
}

func newConfig(authURL, tlsClientCert, tlsClientKey string) (*config, error) {
	cfg := &config{}

	conn, err := db.NewDB(driverName, dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to DB: %w", err)
	}
	cfg.DB = conn

	schm := scheme.GetSchemeFromURL(authURL)

	switch schm {
	case scheme.HTTP, scheme.HTTPS:
		opts := []httpClient.Opt{}
		if tlsClientCert != "" && tlsClientKey != "" {
			opts = append(opts, httpClient.WithTLSClientCert(tlsClientCert, tlsClientKey))
		}

		httpClient, err := httpClient.New(authURL, opts...)
		if err != nil {
			return nil, fmt.Errorf("failed to create HTTP client: %w", err)
		}
		cfg.HTTPClient = httpClient

	case scheme.GRPC:
		opts := []grpcClient.Opt{}
		if tlsClientCert != "" && tlsClientKey != "" {
			opts = append(opts, grpcClient.WithTLSClientCert(tlsClientCert, tlsClientKey))
		} else {
			// Add insecure option explicitly here
			opts = append(opts, grpcClient.WithInsecure())
		}

		grpcClient, err := grpcClient.New(authURL, opts...)
		if err != nil {
			return nil, fmt.Errorf("failed to create gRPC client: %w", err)
		}
		cfg.GRPCClient = grpcClient

	default:
		return nil, fmt.Errorf("unsupported URL scheme: %s", schm)
	}

	return cfg, nil
}
