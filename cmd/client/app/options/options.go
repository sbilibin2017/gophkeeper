package options

import (
	"fmt"
	"net/url"
	"os"

	"github.com/sbilibin2017/gophkeeper/internal/configs"
	"github.com/spf13/cobra"
)

// Options содержит параметры клиента, включая токен, URL сервера и итоговую конфигурацию клиента.
type Options struct {
	Token        string
	ServerURL    string
	ClientConfig *configs.ClientConfig
}

// Opt — функциональный тип опции для конфигурации.
type Opt func(*Options) error

// NewOptions создаёт конфигурацию клиента, включая внутреннюю конфигурацию из internal/configs.
func NewOptions(o ...Opt) (*Options, error) {
	opts := &Options{}

	for _, opt := range o {
		if err := opt(opts); err != nil {
			return nil, err
		}
	}

	if opts.Token == "" {
		opts.Token = os.Getenv("GOPHKEEPER_TOKEN")
	}
	if opts.ServerURL == "" {
		opts.ServerURL = os.Getenv("GOPHKEEPER_SERVER_URL")
	}

	u, err := url.Parse(opts.ServerURL)
	if err != nil {
		return nil, fmt.Errorf("invalid server URL: %w", err)
	}

	clientOpts := []configs.ClientConfigOpt{
		configs.WithToken(opts.Token),
		configs.WithServerURL(opts.ServerURL),
	}

	switch u.Scheme {
	case "http", "https":
		clientOpts = append(clientOpts, configs.WithHTTPClient(opts.ServerURL))
	case "grpc", "grpcs":
		clientOpts = append(clientOpts, configs.WithGRPCClient(opts.ServerURL))
	default:
		return nil, fmt.Errorf("unsupported URL scheme: %s", u.Scheme)
	}

	clientConfig, err := configs.NewClientConfig(clientOpts...)
	if err != nil {
		return nil, err
	}

	opts.ClientConfig = clientConfig
	return opts, nil
}

// WithToken задаёт токен для клиента.
func WithToken(token string) Opt {
	return func(cfg *Options) error {
		cfg.Token = token
		return nil
	}
}

// WithServerURL задаёт URL сервера.
func WithServerURL(serverURL string) Opt {
	return func(cfg *Options) error {
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

// RegisterTokenFlag регистрирует флаг для токена сервера
func RegisterTokenFlag(cmd *cobra.Command, token *string) *cobra.Command {
	cmd.Flags().StringVar(token, "token", "", "Токен авторизации (можно задать через GOPHKEEPER_TOKEN)")
	return cmd
}

// RegisterServerURLFlag регистрирует флаг для URL сервера
func RegisterServerURLFlag(cmd *cobra.Command, serverURL *string) *cobra.Command {
	cmd.Flags().StringVar(serverURL, "server-url", "", "URL сервера (можно задать через GOPHKEEPER_SERVER_URL)")
	return cmd
}

// RegisterInteractiveFlag регистрирует флаг для включения интерактивного режима
func RegisterInteractiveFlag(cmd *cobra.Command, interactive *bool) *cobra.Command {
	cmd.Flags().BoolVar(interactive, "interactive", false, "Включить интерактивный режим ввода")
	return cmd
}
