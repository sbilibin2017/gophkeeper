package clients

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"time"

	"github.com/go-resty/resty/v2"
)

// HTTPClientOption defines a function type for configuring a Resty client.
type HTTPClientOption func(*resty.Client) error

// WithHTTPTLSServerCert добавляет доверенный серверный сертификат (public CA) для TLS.
// certFile — путь к публичному сертификату сервера.
func WithHTTPTLSServerCert(certFile string) HTTPClientOption {
	return func(client *resty.Client) error {
		// Читаем сертификат сервера
		caCert, err := os.ReadFile(certFile)
		if err != nil {
			return fmt.Errorf("failed to read server cert file: %w", err)
		}

		// Создаем новый пул корневых сертификатов и добавляем туда серверный сертификат
		caPool := x509.NewCertPool()
		if !caPool.AppendCertsFromPEM(caCert) {
			return fmt.Errorf("failed to append server cert to pool")
		}

		tlsConfig := &tls.Config{
			RootCAs: caPool,
		}

		client.SetTLSClientConfig(tlsConfig)
		return nil
	}
}

// WithHTTPRetryCount sets the number of retry attempts for HTTP requests.
// Parameters:
// - count: the retry count.
// Returns:
// - a HTTPClientOption function that configures the retry count.
func WithHTTPRetryCount(count int) HTTPClientOption {
	return func(client *resty.Client) error {
		client.SetRetryCount(count)
		return nil
	}
}

// WithHTTPRetryWaitTime sets the minimum wait time between retries.
// Parameters:
// - d: duration to wait before retrying.
// Returns:
// - a HTTPClientOption function that configures the retry wait time.
func WithHTTPRetryWaitTime(d time.Duration) HTTPClientOption {
	return func(client *resty.Client) error {
		client.SetRetryWaitTime(d)
		return nil
	}
}

// WithHTTPRetryMaxWaitTime sets the maximum wait time between retries.
// Parameters:
// - d: maximum duration to wait before retrying.
// Returns:
// - a HTTPClientOption function that configures the max retry wait time.
func WithHTTPRetryMaxWaitTime(d time.Duration) HTTPClientOption {
	return func(client *resty.Client) error {
		client.SetRetryMaxWaitTime(d)
		return nil
	}
}

// NewHTTPClient creates a new Resty client with the base URL and optional configurations.
// It sets default values for retry attempts and wait times before applying custom options.
// Parameters:
// - baseURL: the base URL for the HTTP client.
// - opts: optional HTTPClientOption functions to customize the client.
// Returns:
// - a configured *resty.Client instance.
// - an error if any option fails to apply.
func NewHTTPClient(baseURL string, opts ...HTTPClientOption) (*resty.Client, error) {
	client := resty.New().SetBaseURL(baseURL)

	// Default values
	client.SetRetryCount(3)
	client.SetRetryWaitTime(500 * time.Millisecond)
	client.SetRetryMaxWaitTime(2 * time.Second)

	// Apply custom options which can override defaults
	for _, opt := range opts {
		if err := opt(client); err != nil {
			return nil, err
		}
	}

	return client, nil
}
