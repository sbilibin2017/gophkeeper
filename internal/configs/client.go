package configs

import (
	"encoding/json"

	"github.com/go-resty/resty/v2"
	"github.com/jmoiron/sqlx"
	"github.com/sbilibin2017/gophkeeper/internal/configs/clients"
	"github.com/sbilibin2017/gophkeeper/internal/configs/db"
	"google.golang.org/grpc"
)

// Config хранит конфигурацию клиента, включая HTTP клиент, gRPC клиент и подключение к базе данных.
type ClientConfig struct {
	HTTPClient *resty.Client
	GRPCClient *grpc.ClientConn
	DB         *sqlx.DB
}

// Opt определяет функцию опции конфигурации клиента.
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
func WithClientConfigHTTPClient(baseURL ...string) ClientConfigOpt {
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
func WithClientConfigGRPCClient(addrs ...string) ClientConfigOpt {
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
func WithClientConfigDB(dsn ...string) ClientConfigOpt {
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

// PrepareMetaJSON парсит JSON-строку meta и возвращает её обратно в *string.
// Проверяет корректность JSON, возвращает ошибку, если невалидно.
func PrepareMetaJSON(meta string) (*string, error) {
	if meta == "" {
		return nil, nil
	}

	var parsed map[string]string
	if err := json.Unmarshal([]byte(meta), &parsed); err != nil {
		return nil, err
	}

	b, err := json.Marshal(parsed)
	if err != nil {
		return nil, err
	}

	s := string(b)
	return &s, nil
}
