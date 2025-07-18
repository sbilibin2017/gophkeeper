package clients

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"time"

	"github.com/go-resty/resty/v2"
)

// HTTPClientOption defines a function type for configuring a Resty client.
type HTTPClientOption func(*resty.Client) error

// WithHTTPTLSClientCert добавляет клиентский TLS сертификат и ключ для аутентификации.
// certFile — путь к файлу сертификата клиента.
// keyFile — путь к файлу приватного ключа клиента.
func WithHTTPTLSClientCert(certFile, keyFile string) HTTPClientOption {
	return func(client *resty.Client) error {
		// Загружаем клиентский сертификат и ключ
		cert, err := tls.LoadX509KeyPair(certFile, keyFile)
		if err != nil {
			return fmt.Errorf("failed to load client certificate/key: %w", err)
		}

		// Загружаем системный пул корневых сертификатов
		rootCAs, err := x509.SystemCertPool()
		if err != nil || rootCAs == nil {
			rootCAs = x509.NewCertPool()
		}

		tlsConfig := &tls.Config{
			Certificates: []tls.Certificate{cert},
			RootCAs:      rootCAs,
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

	// Default retry settings
	client.SetRetryCount(3)
	client.SetRetryWaitTime(500 * time.Millisecond)
	client.SetRetryMaxWaitTime(2 * time.Second)

	// Apply user options
	for _, opt := range opts {
		if err := opt(client); err != nil {
			return nil, err
		}
	}

	return client, nil
}
