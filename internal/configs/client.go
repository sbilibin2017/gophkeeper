package configs

import (
	"github.com/go-resty/resty/v2"
	"github.com/jmoiron/sqlx"
	"github.com/sbilibin2017/gophkeeper/internal/configs/clients"
	"github.com/sbilibin2017/gophkeeper/internal/configs/db"
)

// ClientConfig хранит конфигурацию клиента, включая владельца секрета, токен, HTTP клиент и подключение к базе данных.
type ClientConfig struct {
	HTTPClient *resty.Client
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

// WithClient устанавливает базовый URL клиента по первому непустому значению.
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
