package configs

import (
	"errors"
	"net/url"

	"github.com/go-resty/resty/v2"
	"github.com/jmoiron/sqlx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	_ "modernc.org/sqlite"
)

// ClientConfig содержит конфигурации HTTP клиента, gRPC клиента,
// подключения к базе данных и генератора JWT токенов.
type ClientConfig struct {
	HTTPClient *resty.Client    // HTTP клиент
	GRPCClient *grpc.ClientConn // gRPC клиент
	DB         *sqlx.DB         // подключение к базе данных
}

// ClientConfigOpt определяет функцию настройки ClientConfig.
type ClientConfigOpt func(*ClientConfig) error

// NewClientConfig создаёт новую конфигурацию ClientConfig, применяя переданные опции.
// Если какая-либо опция возвращает ошибку, процесс прерывается и ошибка возвращается.
func NewClientConfig(opts ...ClientConfigOpt) (*ClientConfig, error) {
	c := &ClientConfig{}
	for _, opt := range opts {
		if err := opt(c); err != nil {
			return nil, err
		}
	}
	return c, nil
}

// WithDB настраивает SQLite базу данных.
// Использует первый непустой путь из переданных аргументов.
// Возвращает ошибку, если путь не указан или не удалось подключиться к базе.
func WithDB(pathToDB ...string) ClientConfigOpt {
	return func(c *ClientConfig) error {
		var path string
		for _, p := range pathToDB {
			if p != "" {
				path = p
				break
			}
		}
		if path == "" {
			return errors.New("database path not provided")
		}

		db, err := sqlx.Open("sqlite", path)
		if err != nil {
			return err
		}

		if err := db.Ping(); err != nil {
			return err
		}

		c.DB = db
		return nil
	}
}

// WithHTTPClient настраивает HTTP клиента с базовым URL.
// Использует первый непустой URL из переданных аргументов.
// Возвращает ошибку, если URL не указан или имеет неверный формат.
func WithHTTPClient(serverURL ...string) ClientConfigOpt {
	return func(c *ClientConfig) error {
		var urlStr string
		for _, s := range serverURL {
			if s != "" {
				urlStr = s
				break
			}
		}
		if urlStr == "" {
			return errors.New("HTTP server URL not provided")
		}

		parsedURL, err := url.Parse(urlStr)
		if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
			return errors.New("invalid HTTP server URL format")
		}

		c.HTTPClient = resty.New().SetBaseURL(parsedURL.String())
		return nil
	}
}

// WithGRPCClient настраивает gRPC клиента.
// Использует первый непустой URL из переданных аргументов.
// Возвращает ошибку, если URL не указан, имеет неверный формат или не удалось подключиться.
func WithGRPCClient(serverURL ...string) ClientConfigOpt {
	return func(c *ClientConfig) error {
		var addr string
		for _, s := range serverURL {
			if s != "" {
				addr = s
				break
			}
		}

		conn, err := grpc.NewClient(
			addr,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)
		if err != nil {
			return err
		}

		c.GRPCClient = conn
		return nil
	}
}
