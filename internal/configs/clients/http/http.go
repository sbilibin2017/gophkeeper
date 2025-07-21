package http

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"time"

	"github.com/go-resty/resty/v2"
)

// Opt defines a function type to configure resty.Client and potentially return an error.
type Opt func(*resty.Client) error

func New(baseURL string, opts ...Opt) (*resty.Client, error) {
	client := resty.New().SetBaseURL(baseURL)

	for _, opt := range opts {
		if err := opt(client); err != nil {
			return nil, err
		}
	}

	return client, nil
}

// RetryPolicy defines retry configuration for HTTP requests.
type RetryPolicy struct {
	Count   int
	Wait    time.Duration
	MaxWait time.Duration
}

// WithRetryPolicy applies the first valid RetryPolicy in opts.
func WithRetryPolicy(opts ...RetryPolicy) Opt {
	return func(c *resty.Client) error {
		for _, policy := range opts {
			if policy.Count > 0 || policy.Wait > 0 || policy.MaxWait > 0 {
				if policy.Count > 0 {
					c.SetRetryCount(policy.Count)
				}
				if policy.Wait > 0 {
					c.SetRetryWaitTime(policy.Wait)
				}
				if policy.MaxWait > 0 {
					c.SetRetryMaxWaitTime(policy.MaxWait)
				}
				break
			}
		}
		return nil
	}
}

// TLSCert represents a certificate and key file pair.
type TLSCert struct {
	CertFile string
	KeyFile  string
}

// WithTLSCert sets TLS cert/key using the first valid TLSCert in opts.
func WithTLSCert(opts ...TLSCert) Opt {
	return func(c *resty.Client) error {
		for _, certPair := range opts {
			if certPair.CertFile != "" && certPair.KeyFile != "" {
				// Load client cert/key
				cert, err := tls.LoadX509KeyPair(certPair.CertFile, certPair.KeyFile)
				if err != nil {
					return fmt.Errorf("failed to load TLS cert/key: %w", err)
				}

				// Load CA cert (server cert) to RootCAs
				caCertPEM, err := os.ReadFile(certPair.CertFile)
				if err != nil {
					return fmt.Errorf("failed to read cert file: %w", err)
				}

				caCertPool, err := x509.SystemCertPool()
				if err != nil {
					// fallback if SystemCertPool is not available
					caCertPool = x509.NewCertPool()
				}

				if ok := caCertPool.AppendCertsFromPEM(caCertPEM); !ok {
					return fmt.Errorf("failed to append cert to root CA pool")
				}

				c.SetTLSClientConfig(&tls.Config{
					Certificates: []tls.Certificate{cert},
					RootCAs:      caCertPool,
					MinVersion:   tls.VersionTLS12,
				})
				break
			}
		}
		return nil
	}
}

// WithAuthToken sets the Authorization header with a Bearer token.
func WithAuthToken(opts ...string) Opt {
	return func(c *resty.Client) error {
		for _, token := range opts {
			if token != "" {
				c.SetAuthToken(token)
				break
			}
		}
		return nil
	}
}
