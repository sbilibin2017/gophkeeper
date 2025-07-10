package config

import (
	"fmt"
	"net/url"
	"os"

	"github.com/sbilibin2017/gophkeeper/internal/configs"
)

// Config содержит параметры клиента, включая токен, URL сервера и итоговую конфигурацию клиента.
type Config struct {
	Token        string
	ServerURL    string
	ClientConfig *configs.ClientConfig
}

// ConfigOpt — функциональный тип опции для конфигурации.
type ConfigOpt func(*Config) error

// NewConfig создаёт конфигурацию клиента, включая внутреннюю конфигурацию из internal/configs.
func NewConfig(opts ...ConfigOpt) (*Config, error) {
	cfg := &Config{}

	for _, opt := range opts {
		if err := opt(cfg); err != nil {
			return nil, err
		}
	}

	if cfg.Token == "" {
		cfg.Token = os.Getenv("GOPHKEEPER_TOKEN")
	}
	if cfg.ServerURL == "" {
		cfg.ServerURL = os.Getenv("GOPHKEEPER_SERVER_URL")
	}

	u, err := url.Parse(cfg.ServerURL)
	if err != nil {
		return nil, fmt.Errorf("invalid server URL: %w", err)
	}

	clientOpts := []configs.ClientConfigOpt{
		configs.WithToken(cfg.Token),
		configs.WithServerURL(cfg.ServerURL),
	}

	switch u.Scheme {
	case "http", "https":
		clientOpts = append(clientOpts, configs.WithHTTPClient(cfg.ServerURL))
	case "grpc", "grpcs":
		clientOpts = append(clientOpts, configs.WithGRPCClient(cfg.ServerURL))
	default:
		return nil, fmt.Errorf("unsupported URL scheme: %s", u.Scheme)
	}

	clientConfig, err := configs.NewClientConfig(clientOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to build client config: %w", err)
	}

	cfg.ClientConfig = clientConfig
	return cfg, nil
}

// WithToken задаёт токен для клиента.
func WithToken(token string) ConfigOpt {
	return func(cfg *Config) error {
		cfg.Token = token
		return nil
	}
}

// WithServerURL задаёт URL сервера.
func WithServerURL(serverURL string) ConfigOpt {
	return func(cfg *Config) error {
		cfg.ServerURL = serverURL
		return nil
	}
}

// SetToken устанавливает токен в переменную окружения и обновляет конфигурацию.
func SetToken(token string) error {
	if token == "" {
		return fmt.Errorf("token is empty")
	}
	if err := os.Setenv("GOPHKEEPER_TOKEN", token); err != nil {
		return fmt.Errorf("failed to set environment variable GOPHKEEPER_TOKEN: %w", err)
	}
	return nil
}

// SetServerURL устанавливает URL сервера в переменную окружения и обновляет конфигурацию.
func SetServerURL(serverURL string) error {
	if serverURL == "" {
		return fmt.Errorf("server URL is empty")
	}
	if err := os.Setenv("GOPHKEEPER_SERVER_URL", serverURL); err != nil {
		return fmt.Errorf("failed to set environment variable GOPHKEEPER_SERVER_URL: %w", err)
	}
	return nil
}
