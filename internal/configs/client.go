package configs

import (
	"errors"
	"net/url"
	"os"

	"github.com/go-resty/resty/v2"
	"github.com/jmoiron/sqlx"
	"github.com/sbilibin2017/gophkeeper/internal/configs/jwt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	_ "modernc.org/sqlite"
)

// ClientConfig содержит конфигурации HTTP клиента, gRPC клиента,
// подключения к базе данных и генератора JWT токенов.
type ClientConfig struct {
	HTTPClient   *resty.Client                         // HTTP клиент
	GRPCClient   *grpc.ClientConn                      // gRPC клиент
	DB           *sqlx.DB                              // подключение к базе данных
	JWTGenerator func(username string) (string, error) // функция генерации JWT токена по имени пользователя
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

// WithJWTGenerator задаёт функцию генерации JWT токена.
// Использует первый непустой секретный ключ из переданных аргументов.
// Возвращает ошибку, если секретный ключ не указан.
func WithJWTGenerator(secretKey ...string) ClientConfigOpt {
	return func(c *ClientConfig) error {
		var secret string
		for _, s := range secretKey {
			if s != "" {
				secret = s
				break
			}
		}
		if secret == "" {
			return errors.New("JWT secret key not provided")
		}

		c.JWTGenerator = func(username string) (string, error) {
			return jwt.GenerateToken(username, secret)
		}
		return nil
	}
}

// SetServerURLToEnv устанавливает переменную окружения GOPHKEEPER_SERVER_URL с переданным URL сервера.
// Возвращает ошибку, если установка переменной окружения не удалась.
func SetServerURLToEnv(serverURL string) error {
	return os.Setenv("GOPHKEEPER_SERVER_URL", serverURL)
}

// GetServerURLFromEnv возвращает значение переменной окружения GOPHKEEPER_SERVER_URL.
// Если переменная не установлена, возвращает пустую строку.
func GetServerURLFromEnv() string {
	return os.Getenv("GOPHKEEPER_SERVER_URL")
}

// SetTokenToEnv устанавливает переменную окружения GOPHKEEPER_TOKEN с переданным токеном.
// Возвращает ошибку, если установка переменной окружения не удалась.
func SetTokenToEnv(token string) error {
	return os.Setenv("GOPHKEEPER_TOKEN", token)
}

// GetTokenFromEnv возвращает значение переменной окружения GOPHKEEPER_TOKEN.
// Если переменная не установлена, возвращает пустую строку.
func GetTokenFromEnv() string {
	return os.Getenv("GOPHKEEPER_TOKEN")
}
