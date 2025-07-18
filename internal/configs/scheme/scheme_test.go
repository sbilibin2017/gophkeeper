package scheme

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetSchemeFromURL(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected string
	}{
		{
			name:     "HTTP scheme",
			url:      "http://example.com",
			expected: HTTP,
		},
		{
			name:     "HTTPS scheme",
			url:      "https://example.com",
			expected: HTTPS,
		},
		{
			name:     "gRPC scheme",
			url:      "grpc://service.local",
			expected: GRPC,
		},
		{
			name:     "Unknown scheme",
			url:      "ftp://example.com",
			expected: "",
		},
		{
			name:     "Empty string",
			url:      "",
			expected: "",
		},
		{
			name:     "Partial match",
			url:      "httpx://malformed",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := GetSchemeFromURL(tt.url)
			assert.Equal(t, tt.expected, actual)
		})
	}
}
