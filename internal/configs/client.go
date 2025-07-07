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

// ClientConfig holds the client configuration (HTTP/gRPC) and encoding functions (HMAC/RSA).
type ClientConfig struct {
	HTTPClient  *resty.Client
	GRPCClient  *grpc.ClientConn
	HMACEncoder func([]byte) []byte
	RSAEncoder  func([]byte) ([]byte, error)
}

// ClientConfigOpt defines a function that modifies ClientConfig and may return an error.
type ClientConfigOpt func(*ClientConfig) error

// NewClientConfig creates a new instance of ClientConfig and applies the given options.
// Returns an error if any of the options fail.
func NewClientConfig(opts ...ClientConfigOpt) (*ClientConfig, error) {
	c := &ClientConfig{}
	for _, opt := range opts {
		if err := opt(c); err != nil {
			return nil, err
		}
	}
	return c, nil
}

// WithClient configures the client based on the server URL scheme.
// Supports HTTP(S) and gRPC protocols.
func WithClient(serverURL string) ClientConfigOpt {
	return func(c *ClientConfig) error {
		parsed, err := url.Parse(serverURL)
		if err != nil {
			return fmt.Errorf("invalid server URL: %w", err)
		}

		switch parsed.Scheme {
		case "http", "https":
			c.HTTPClient = resty.New().
				SetBaseURL(serverURL).
				SetHeader("Accept", "application/json")
			return nil

		case "grpc":
			grpcClient, err := grpc.NewClient(
				parsed.Host,
				grpc.WithTransportCredentials(insecure.NewCredentials()),
			)
			if err != nil {
				return fmt.Errorf("failed to create gRPC client: %w", err)
			}

			c.GRPCClient = grpcClient
			return nil

		default:
			return fmt.Errorf("unsupported URL scheme: %s", parsed.Scheme)
		}
	}
}

// WithHMACEncoder sets an HMAC-SHA256 encoding function using the provided key.
// Returns an error if the key is empty.
func WithHMACEncoder(key string) ClientConfigOpt {
	return func(c *ClientConfig) error {
		if key == "" {
			return nil
		}
		keyBytes := []byte(key)

		c.HMACEncoder = func(data []byte) []byte {
			mac := hmac.New(sha256.New, keyBytes)
			mac.Write(data)
			return mac.Sum(nil)
		}
		return nil
	}
}

// WithRSAEncoder sets an RSA encryption function using a public key read from a PEM file.
// Expects a PEM block of type "PUBLIC KEY".
func WithRSAEncoder(publicKeyPath string) ClientConfigOpt {
	return func(c *ClientConfig) error {
		if publicKeyPath == "" {
			return nil
		}

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

		c.RSAEncoder = func(data []byte) ([]byte, error) {
			return rsa.EncryptPKCS1v15(rand.Reader, rsaPub, data)
		}

		return nil
	}
}
