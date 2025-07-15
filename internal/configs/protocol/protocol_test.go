package protocol

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetProtocolFromURL(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected string
	}{
		{"HTTP prefix", "http://example.com", HTTP},
		{"HTTPS prefix", "https://example.com", HTTPS},
		{"gRPC prefix", "grpc://service", GRPC},
		{"No prefix", "ftp://example.com", ""},
		{"Empty string", "", ""},
		{"Prefix in middle", "example.com/http://", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetProtocolFromURL(tt.url)
			assert.Equal(t, tt.expected, got)
		})
	}
}
