package configs

import (
	"net/url"

	"github.com/go-resty/resty/v2"
	"github.com/jmoiron/sqlx"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	_ "modernc.org/sqlite"
)

// ClientConfig содержит конфигурацию для HTTP клиента, gRPC клиента и подключения к базе данных.
type ClientConfig struct {
	HTTPClient *resty.Client
	GRPCClient *grpc.ClientConn
	DB         *sqlx.DB
}

// ClientConfigOpt определяет тип функции опции для настройки ClientConfig.
type ClientConfigOpt func(*ClientConfig) error

// NewClientConfig создаёт новый ClientConfig, применяя переданные опции.
// Возвращает ошибку, если какая-либо из опций вернула ошибку.
func NewClientConfig(opts ...ClientConfigOpt) (*ClientConfig, error) {
	c := &ClientConfig{}
	for _, opt := range opts {
		if err := opt(c); err != nil {
			return nil, err
		}
	}
	return c, nil
}

// WithHTTPClient настраивает ClientConfig для использования HTTP клиента
// с базовым URL, заданным в serverURL.
func WithHTTPClient(serverURL string) ClientConfigOpt {
	return func(c *ClientConfig) error {
		client := resty.New().SetBaseURL(serverURL)
		c.HTTPClient = client
		return nil
	}
}

// WithGRPCClient настраивает ClientConfig для использования gRPC клиента
func WithGRPCClient(serverURL string) ClientConfigOpt {
	return func(c *ClientConfig) error {
		parsed, err := url.Parse(serverURL)
		if err != nil {
			return err
		}

		client, err := grpc.NewClient(
			parsed.Host,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)
		if err != nil {
			return err
		}

		c.GRPCClient = client
		return nil
	}
}
