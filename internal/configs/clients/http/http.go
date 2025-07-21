package http

import (
	"crypto/tls"
	"fmt"
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
				cert, err := tls.LoadX509KeyPair(certPair.CertFile, certPair.KeyFile)
				if err != nil {
					return fmt.Errorf("failed to load TLS cert/key: %w", err)
				}
				c.SetTLSClientConfig(&tls.Config{Certificates: []tls.Certificate{cert}})
				break
			}
		}
		return nil
	}
}

// WithToken sets the Authorization header with a Bearer token.
func WithToken(opts ...string) Opt {
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
