package protocol

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetProtocol_ParseError(t *testing.T) {
	// This input is intentionally malformed to trigger url.Parse error
	invalidURL := "http://[::1" // missing closing bracket

	protocol, err := GetProtocol(invalidURL)

	assert.Error(t, err)
	assert.Empty(t, protocol)
}

func TestGetProtocol_ValidProtocols(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"HTTP protocol", "http://example.com", HTTP},
		{"HTTPS protocol", "https://secure.com", HTTPS},
		{"GRPC protocol", "grpc://service", GRPC},
		{"Upper case scheme", "HTTPS://secure.com", HTTPS},
		{"Scheme with spaces", "  http://example.com  ", HTTP},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			protocol, err := GetProtocol(tt.input)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, protocol)
		})
	}
}

func TestGetProtocol_InvalidInputs(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"No scheme", "example.com"},
		{"Unsupported scheme", "ftp://example.com"},
		{"Empty string", ""},
		{"Whitespace only", "   "},
		{"Malformed URL", "http//missing-colon.com"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			protocol, err := GetProtocol(tt.input)
			assert.Error(t, err)
			assert.Empty(t, protocol)
		})
	}
}
