package configs

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net/url"
	"os"

	"github.com/go-resty/resty/v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// ClientConfig содержит конфигурацию клиента (HTTP/gRPC) и функции кодирования (HMAC/RSA).
type ClientConfig struct {
	httpClient  *resty.Client
	grpcClient  *grpc.ClientConn
	hmacEncoder func([]byte) []byte
	rsaEncoder  func([]byte) ([]byte, error)
}

// ClientConfigOpt определяет функцию, модифицирующую ClientConfig и, возможно, возвращающую ошибку.
type ClientConfigOpt func(*ClientConfig) error

// NewClientConfig создает новый экземпляр ClientConfig и применяет переданные опции.
// Возвращает ошибку, если какая-либо из опций завершилась неудачно.
func NewClientConfig(opts ...ClientConfigOpt) (*ClientConfig, error) {
	c := &ClientConfig{}
	for _, opt := range opts {
		if err := opt(c); err != nil {
			return nil, err
		}
	}
	return c, nil
}

// WithClient инициализирует HTTP- или gRPC-клиент в зависимости от схемы URL.
// Поддерживаемые схемы: "http", "https", "grpc".
func WithClient(serverURL string) ClientConfigOpt {
	return func(c *ClientConfig) error {
		parsed, err := url.Parse(serverURL)
		if err != nil {
			return fmt.Errorf("invalid server URL: %w", err)
		}

		switch parsed.Scheme {
		case "http", "https":
			c.httpClient = resty.New().
				SetBaseURL(serverURL).
				SetHeader("Accept", "application/json")
			return nil

		case "grpc":
			clientConn, err := grpc.NewClient(
				serverURL,
				grpc.WithTransportCredentials(insecure.NewCredentials()),
			)
			if err != nil {
				return fmt.Errorf("failed to connect to gRPC server: %w", err)
			}
			c.grpcClient = clientConn
			return nil

		default:
			return fmt.Errorf("unsupported URL scheme: %s", parsed.Scheme)
		}
	}
}

// WithHMACEncoder устанавливает функцию кодирования HMAC-SHA256 с использованием заданного ключа.
// Возвращает ошибку, если ключ пустой.
func WithHMACEncoder(key string) ClientConfigOpt {
	return func(c *ClientConfig) error {
		if key == "" {
			return fmt.Errorf("HMAC key cannot be empty")
		}
		keyBytes := []byte(key)

		c.hmacEncoder = func(data []byte) []byte {
			mac := hmac.New(sha256.New, keyBytes)
			mac.Write(data)
			return mac.Sum(nil)
		}
		return nil
	}
}

// WithRSAEncoder устанавливает функцию RSA-шифрования, используя открытый ключ из PEM-файла.
// Ожидается PEM-блок с типом "PUBLIC KEY".
func WithRSAEncoder(publicKeyPath string) ClientConfigOpt {
	return func(c *ClientConfig) error {
		data, err := os.ReadFile(publicKeyPath)
		if err != nil {
			return fmt.Errorf("failed to read RSA public key file: %w", err)
		}

		block, _ := pem.Decode(data)
		if block == nil || block.Type != "PUBLIC KEY" {
			return fmt.Errorf("invalid PEM format or missing PUBLIC KEY block")
		}

		pub, err := x509.ParsePKIXPublicKey(block.Bytes)
		if err != nil {
			return fmt.Errorf("failed to parse RSA public key: %w", err)
		}

		rsaPub, ok := pub.(*rsa.PublicKey)
		if !ok {
			return fmt.Errorf("provided key is not a valid RSA public key")
		}

		c.rsaEncoder = func(data []byte) ([]byte, error) {
			return rsa.EncryptPKCS1v15(rand.Reader, rsaPub, data)
		}

		return nil
	}
}
