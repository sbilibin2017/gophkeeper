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

// ClientConfig holds HTTP/gRPC clients and HMAC/RSA encoding functions.
type ClientConfig struct {
	httpClient  *resty.Client
	grpcClient  *grpc.ClientConn
	hmacEncoder func([]byte) []byte
	rsaEncoder  func([]byte) ([]byte, error)
}

// ClientConfigOpt defines a function that modifies a ClientConfig and may return an error.
type ClientConfigOpt func(*ClientConfig) error

// NewClientConfig creates a new ClientConfig instance and applies the provided options.
// Returns an error if any option fails.
func NewClientConfig(opts ...ClientConfigOpt) (*ClientConfig, error) {
	c := &ClientConfig{}
	for _, opt := range opts {
		if err := opt(c); err != nil {
			return nil, err
		}
	}
	return c, nil
}

// WithClient initializes either an HTTP or gRPC client depending on the URL scheme.
// Supported schemes: "http", "https", "grpc".
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

// WithHMACEncoder sets up an HMAC-SHA256 encoding function using the provided secret key.
// Returns an error if the key is empty.
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

// WithRSAEncoder sets up RSA encryption using a public key from a PEM file.
// Expects a PEM block of type "PUBLIC KEY".
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
