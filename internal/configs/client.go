package configs

import (
	"fmt"
	"net/url"

	"github.com/go-resty/resty/v2"
	"github.com/jmoiron/sqlx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	_ "modernc.org/sqlite"
)

// ClientConfig содержит конфигурацию для HTTP клиента, gRPC клиента и подключения к базе данных.
type ClientConfig struct {
	Token      string           // JWT токен для аутентификации
	ServerURL  string           // URL сервера
	HTTPClient *resty.Client    // HTTP клиент
	GRPCClient *grpc.ClientConn // gRPC клиент
	DB         *sqlx.DB         // Подключение к базе данных
}

// ClientConfigOpt определяет функцию для настройки ClientConfig.
type ClientConfigOpt func(*ClientConfig) error

// NewClientConfig создаёт новую конфигурацию ClientConfig, применяя переданные опции.
// Возвращает ошибку, если какая-либо опция завершилась неудачей.
func NewClientConfig(opts ...ClientConfigOpt) (*ClientConfig, error) {
	c := &ClientConfig{}
	for _, opt := range opts {
		if err := opt(c); err != nil {
			return nil, err
		}
	}
	return c, nil
}

// WithToken задаёт JWT токен для ClientConfig.
// Если токен пустой, опция игнорируется.
func WithToken(token string) ClientConfigOpt {
	return func(c *ClientConfig) error {
		if token == "" {
			return nil
		}
		c.Token = token
		return nil
	}
}

// WithServerURL задаёт URL сервера для ClientConfig.
// Если URL пустой, опция игнорируется.
func WithServerURL(serverURL string) ClientConfigOpt {
	return func(c *ClientConfig) error {
		if serverURL == "" {
			return nil
		}
		c.ServerURL = serverURL
		return nil
	}
}

// WithDB настраивает ClientConfig для использования SQLite базы данных,
// расположенной по указанному пути.
// Возвращает ошибку при невозможности открыть или подключиться к базе.
func WithDB(pathToDB string) ClientConfigOpt {
	return func(c *ClientConfig) error {
		db, err := sqlx.Open("sqlite", pathToDB)
		if err != nil {
			return fmt.Errorf("не удалось открыть базу данных: %w", err)
		}

		if err := db.Ping(); err != nil {
			return fmt.Errorf("не удалось подключиться к базе данных: %w", err)
		}

		c.DB = db
		return nil
	}
}

// WithHTTPClient настраивает ClientConfig для использования HTTP клиента,
// базовый URL которого задан параметром serverURL.
// Проверяет корректность URL и возвращает ошибку при неверном формате.
func WithHTTPClient(serverURL string) ClientConfigOpt {
	return func(c *ClientConfig) error {
		parsedURL, err := url.Parse(serverURL)
		if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
			return fmt.Errorf("некорректный serverURL для HTTP клиента: %s", serverURL)
		}

		c.HTTPClient = resty.New().SetBaseURL(parsedURL.String())
		return nil
	}
}

// WithGRPCClient настраивает ClientConfig для использования gRPC клиента.
// Проверяет корректность serverURL и возвращает ошибку при неправильном формате или сбое подключения.
func WithGRPCClient(serverURL string) ClientConfigOpt {
	return func(c *ClientConfig) error {
		parsedURL, err := url.Parse(serverURL)
		if err != nil || parsedURL.Host == "" {
			return fmt.Errorf("некорректный serverURL для gRPC клиента: %s", serverURL)
		}

		conn, err := grpc.NewClient(parsedURL.Host, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			return fmt.Errorf("не удалось подключиться к gRPC серверу: %w", err)
		}

		c.GRPCClient = conn
		return nil
	}
}
