package configs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewClientConfig_WithOptions(t *testing.T) {
	cfg := NewClientConfig(
		WithServerURL("https://example.com"),
		WithRSAPublicKeyPath("/path/to/key.pub"),
		WithHMACKey("supersecretkey"),
	)

	assert.NotNil(t, cfg)
	assert.Equal(t, "https://example.com", cfg.ServerURL)
	assert.Equal(t, "/path/to/key.pub", cfg.RSAPublicKeyPath)
	assert.Equal(t, "supersecretkey", cfg.HMACKey)
}

func TestWithServerURL(t *testing.T) {
	cfg := &ClientConfig{}
	opt := WithServerURL("http://localhost:8080")
	opt(cfg)

	assert.Equal(t, "http://localhost:8080", cfg.ServerURL)
}

func TestWithRSAPublicKeyPath(t *testing.T) {
	cfg := &ClientConfig{}
	opt := WithRSAPublicKeyPath("/my/key.pub")
	opt(cfg)

	assert.Equal(t, "/my/key.pub", cfg.RSAPublicKeyPath)
}

func TestWithHMACKey(t *testing.T) {
	cfg := &ClientConfig{}
	opt := WithHMACKey("my-hmac-key")
	opt(cfg)

	assert.Equal(t, "my-hmac-key", cfg.HMACKey)
}
