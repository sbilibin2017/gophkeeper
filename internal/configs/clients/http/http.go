package http

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"time"

	"github.com/go-resty/resty/v2"
)

func New(baseURL string, opts ...Opt) (*resty.Client, error) {
	client := resty.New().SetBaseURL(baseURL)
	for _, opt := range opts {
		if err := opt(client); err != nil {
			return nil, err
		}
	}
	return client, nil
}

// Opt defines a function type for configuring a Resty client.
type Opt func(*resty.Client) error

func WithTLSClientCert(certFile string, keyFile string) Opt {
	return func(client *resty.Client) error {
		if certFile == "" || keyFile == "" {
			return fmt.Errorf("certFile and keyFile must not be empty")
		}

		cert, err := tls.LoadX509KeyPair(certFile, keyFile)
		if err != nil {
			return fmt.Errorf("failed to load client certificate/key: %w", err)
		}

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

func WithRetryCount(count ...int) Opt {
	return func(client *resty.Client) error {
		for _, c := range count {
			if c != 0 {
				client.SetRetryCount(c)
				break
			}
		}
		return nil
	}
}

func WithRetryWaitTime(d ...time.Duration) Opt {
	return func(client *resty.Client) error {
		for _, dur := range d {
			if dur != 0 {
				client.SetRetryWaitTime(dur)
				break
			}
		}
		return nil
	}
}

func WithRetryMaxWaitTime(d ...time.Duration) Opt {
	return func(client *resty.Client) error {
		for _, dur := range d {
			if dur != 0 {
				client.SetRetryMaxWaitTime(dur)
				break
			}
		}
		return nil
	}
}

func WithToken(token ...string) Opt {
	return func(client *resty.Client) error {
		for _, t := range token {
			if t != "" {
				client.OnBeforeRequest(func(c *resty.Client, r *resty.Request) error {
					r.SetHeader("Authorization", "Bearer "+t)
					return nil
				})
				break
			}
		}
		return nil
	}
}
