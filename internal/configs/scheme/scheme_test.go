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
			url:      "https://secure.example.com",
			expected: HTTPS,
		},
		{
			name:     "GRPC scheme",
			url:      "grpc://service.local",
			expected: GRPC,
		},
		{
			name:     "Unknown scheme",
			url:      "ftp://example.com",
			expected: "",
		},
		{
			name:     "Empty URL",
			url:      "",
			expected: "",
		},
		{
			name:     "URL with no scheme",
			url:      "example.com",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetSchemeFromURL(tt.url)
			assert.Equal(t, tt.expected, result)
		})
	}
}
