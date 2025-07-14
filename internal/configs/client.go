package configs

import (
	"github.com/go-resty/resty/v2"
	"github.com/jmoiron/sqlx"
	"github.com/sbilibin2017/gophkeeper/internal/configs/clients"
	"github.com/sbilibin2017/gophkeeper/internal/configs/db"
	"google.golang.org/grpc"
)

// ClientConfig хранит конфигурацию клиента, включая HTTP клиент, gRPC клиент и подключение к базе данных.
type ClientConfig struct {
	HTTPClient *resty.Client
	GRPCClient *grpc.ClientConn
	DB         *sqlx.DB
}

// ClientConfigOpt определяет функцию опции конфигурации клиента.
type ClientConfigOpt func(*ClientConfig) error

// NewClientConfig создает новый ClientConfig и применяет к нему переданные опции.
func NewClientConfig(opts ...ClientConfigOpt) (*ClientConfig, error) {
	cfg := &ClientConfig{}
	for _, opt := range opts {
		if err := opt(cfg); err != nil {
			return nil, err
		}
	}
	return cfg, nil
}

// WithHTTPClient устанавливает базовый URL клиента по первому непустому значению.
func WithHTTPClient(baseURL ...string) ClientConfigOpt {
	return func(cfg *ClientConfig) error {
		for _, v := range baseURL {
			if v != "" {
				cfg.HTTPClient = clients.NewHTTPClient(v)
				break
			}
		}
		return nil
	}
}

// WithGRPCClient подключает gRPC клиент по первому непустому адресу.
func WithGRPCClient(addrs ...string) ClientConfigOpt {
	return func(cfg *ClientConfig) error {
		for _, addr := range addrs {
			if addr != "" {
				conn, err := clients.NewGRPCClient(addr)
				if err != nil {
					return err
				}
				cfg.GRPCClient = conn
				return nil
			}
		}
		return nil
	}
}

// WithDB подключает базу данных по первому непустому dsn.
func WithDB(dsn ...string) ClientConfigOpt {
	return func(cfg *ClientConfig) error {
		for _, v := range dsn {
			if v != "" {
				conn, err := db.NewDB("sqlite", v)
				if err != nil {
					return err
				}
				cfg.DB = conn
				return nil
			}
		}
		return nil
	}
}
