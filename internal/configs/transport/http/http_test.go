package http

import (
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
)

func TestNew_BaseURL(t *testing.T) {
	tests := []struct {
		name     string
		baseURL  string
		expected string
	}{
		{"with http", "http://example.com", "http://example.com"},
		{"with https", "https://example.com", "https://example.com"},
		{"without scheme", "example.com", "http://example.com"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := New(tt.baseURL)
			assert.Equal(t, tt.expected, client.BaseURL)
		})
	}
}

func TestWithRetryPolicy(t *testing.T) {
	tests := []struct {
		name        string
		policy      RetryPolicy
		wantCount   int
		wantWait    time.Duration
		wantMaxWait time.Duration
	}{
		{"all zero", RetryPolicy{}, 0, 0, 0},
		{"only count", RetryPolicy{Count: 3}, 3, 100 * time.Millisecond, 2 * time.Second}, // Resty дефолтные значения
		{"count and wait", RetryPolicy{Count: 2, Wait: 100 * time.Millisecond}, 2, 100 * time.Millisecond, 2 * time.Second},
		{"all set", RetryPolicy{Count: 5, Wait: 50 * time.Millisecond, MaxWait: 500 * time.Millisecond}, 5, 50 * time.Millisecond, 500 * time.Millisecond},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := resty.New()
			opt := WithRetryPolicy(tt.policy)
			opt(client)

			assert.Equal(t, tt.wantCount, client.RetryCount)
			assert.Equal(t, tt.wantWait, client.RetryWaitTime)
			assert.Equal(t, tt.wantMaxWait, client.RetryMaxWaitTime)
		})
	}
}

func TestNew(t *testing.T) {
	// Опция для теста
	testOpt := func(c *resty.Client) {
		c.SetTimeout(500 * time.Millisecond)
	}

	tests := []struct {
		name        string
		baseURL     string
		opts        []Opt
		wantBase    string
		wantTimeout time.Duration
	}{
		{"with http", "http://example.com", nil, "http://example.com", 0},
		{"with https", "https://example.com", nil, "https://example.com", 0},
		{"without scheme", "example.com", nil, "http://example.com", 0},
		{"with option", "example.com", []Opt{testOpt}, "http://example.com", 500 * time.Millisecond},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := New(tt.baseURL, tt.opts...)
			assert.Equal(t, tt.wantBase, client.BaseURL)

		})
	}
}
