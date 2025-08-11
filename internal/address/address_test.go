package address

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	tests := []struct {
		input          string
		expectedScheme string
		expectedAddr   string
	}{
		{"http://localhost:8080", SchemeHTTP, "localhost:8080"},
		{"https://example.com", SchemeHTTPS, "example.com"},
		{"grpc://127.0.0.1:50051", SchemeGRPC, "127.0.0.1:50051"},
		{"localhost:8080", SchemeHTTP, "localhost:8080"},       // default scheme
		{"ftp://example.com", SchemeHTTP, "ftp://example.com"}, // unsupported scheme treated as default http + full addr
		{"", SchemeHTTP, ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			addr := New(tt.input)
			assert.Equal(t, tt.expectedScheme, addr.Scheme)
			assert.Equal(t, tt.expectedAddr, addr.Address)
		})
	}
}

func TestString(t *testing.T) {
	addr := Address{
		Scheme:  SchemeGRPC,
		Address: "localhost:50051",
	}
	assert.Equal(t, "grpc://localhost:50051", addr.String())
}
