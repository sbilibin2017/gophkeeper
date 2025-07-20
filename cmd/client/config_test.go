package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func cleanupDB(t *testing.T) {
	err := os.Remove("client.db")
	if err != nil && !os.IsNotExist(err) {
		t.Fatalf("failed to clean up client.db: %v", err)
	}
}

func TestNewConfig_UnsupportedScheme(t *testing.T) {
	defer cleanupDB(t)

	_, err := newConfig("ftp://example.com", "", "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported URL scheme")
}

func TestNewConfig_HTTP_NoTLS(t *testing.T) {
	defer cleanupDB(t)

	cfg, err := newConfig("http://example.com", "", "")
	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.NotNil(t, cfg.DB)
	assert.NotNil(t, cfg.HTTPClient)
	assert.Nil(t, cfg.GRPCClient)
}

func TestNewConfig_HTTPS_WithTLS(t *testing.T) {
	defer cleanupDB(t)

	cert := ""
	key := ""

	cfg, err := newConfig("https://example.com", cert, key)
	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.NotNil(t, cfg.DB)
	assert.NotNil(t, cfg.HTTPClient)
	assert.Nil(t, cfg.GRPCClient)
}

func TestNewConfig_GRPC_NoTLS(t *testing.T) {
	defer cleanupDB(t)

	cfg, err := newConfig("grpc://localhost:50051", "", "")
	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.NotNil(t, cfg.DB)
	assert.Nil(t, cfg.HTTPClient)
	assert.NotNil(t, cfg.GRPCClient)
}

func TestNewConfig_GRPC_WithTLS(t *testing.T) {
	defer cleanupDB(t)

	cert := ""
	key := ""

	cfg, err := newConfig("grpc://localhost:50051", cert, key)
	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.NotNil(t, cfg.DB)
	assert.Nil(t, cfg.HTTPClient)
	assert.NotNil(t, cfg.GRPCClient)
}
