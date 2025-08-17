package address

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantAddr Address
	}{
		{
			name:  "HTTP scheme explicit",
			input: "http://example.com:8080",
			wantAddr: Address{
				Scheme:  SchemeHTTP,
				Address: "example.com:8080",
			},
		},
		{
			name:  "HTTPS scheme explicit",
			input: "https://secure.com",
			wantAddr: Address{
				Scheme:  SchemeHTTPS,
				Address: "secure.com",
			},
		},
		{
			name:  "GRPC scheme explicit",
			input: "grpc://localhost:50051",
			wantAddr: Address{
				Scheme:  SchemeGRPC,
				Address: "localhost:50051",
			},
		},
		{
			name:  "No scheme defaults to HTTP",
			input: "noscheme.com:1234",
			wantAddr: Address{
				Scheme:  SchemeHTTP,
				Address: "noscheme.com:1234",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := New(tt.input)
			assert.Equal(t, tt.wantAddr, got)
			assert.Equal(t, tt.wantAddr.Scheme+"://"+tt.wantAddr.Address, got.String())
		})
	}
}

func TestString(t *testing.T) {
	addr := Address{
		Scheme:  SchemeHTTPS,
		Address: "example.com:443",
	}
	assert.Equal(t, "https://example.com:443", addr.String())
}
