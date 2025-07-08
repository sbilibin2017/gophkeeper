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
	"github.com/jmoiron/sqlx"
	"github.com/sbilibin2017/gophkeeper/internal/configs/db"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type ClientConfig struct {
	HTTPClient *resty.Client
	GRPCClient *grpc.ClientConn
	Encoders   []func(data []byte) ([]byte, error)
	DB         *sqlx.DB
}

type ClientConfigOpt func(*ClientConfig) error

func NewClientConfig(opts ...ClientConfigOpt) (*ClientConfig, error) {
	c := &ClientConfig{}
	for _, opt := range opts {
		if err := opt(c); err != nil {
			return nil, err
		}
	}
	return c, nil
}

func WithDB() ClientConfigOpt {
	return func(c *ClientConfig) error {
		db, err := db.NewDB()
		if err != nil {
			return err
		}
		c.DB = db
		return nil
	}
}

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

func WithHMACEncoder(key string) ClientConfigOpt {
	return func(c *ClientConfig) error {
		if key == "" {
			return nil
		}
		keyBytes := []byte(key)

		hmacEnc := func(data []byte) ([]byte, error) {
			mac := hmac.New(sha256.New, keyBytes)
			mac.Write(data)
			return mac.Sum(nil), nil
		}

		c.Encoders = append(c.Encoders, hmacEnc)
		return nil
	}
}

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

		rsaEnc := func(data []byte) ([]byte, error) {
			return rsa.EncryptPKCS1v15(rand.Reader, rsaPub, data)
		}

		c.Encoders = append(c.Encoders, rsaEnc)
		return nil
	}
}
