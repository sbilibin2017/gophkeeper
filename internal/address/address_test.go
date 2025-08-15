package address

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantScheme string
		wantAddr   string
		wantErr    error
	}{
		{
			name:       "HTTP full address",
			input:      "http://example.com:8080",
			wantScheme: SchemeHTTP,
			wantAddr:   "example.com:8080",
		},
		{
			name:       "HTTPS full address",
			input:      "https://example.com:443",
			wantScheme: SchemeHTTPS,
			wantAddr:   "example.com:443",
		},
		{
			name:       "GRPC full address",
			input:      "grpc://localhost:50051",
			wantScheme: SchemeGRPC,
			wantAddr:   "localhost:50051",
		},
		{
			name:       "Default scheme http",
			input:      "localhost:8080",
			wantScheme: SchemeHTTP,
			wantAddr:   "localhost:8080",
		},
		{
			name:       "Only port",
			input:      ":9090",
			wantScheme: SchemeHTTP,
			wantAddr:   "0.0.0.0:9090",
		},
		{
			name:       "Empty input",
			input:      "",
			wantScheme: SchemeHTTP,
			wantAddr:   "0.0.0.0",
		},
		{
			name:    "Unsupported scheme",
			input:   "ftp://example.com",
			wantErr: ErrUnsupportedScheme,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			addr, err := New(tt.input)
			if tt.wantErr != nil {
				assert.Equal(t, tt.wantErr, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.wantScheme, addr.Scheme)
			assert.Equal(t, tt.wantAddr, addr.Address)
		})
	}
}

func TestAddress_String(t *testing.T) {
	tests := []struct {
		name string
		addr Address
		want string
	}{
		{
			name: "HTTP address",
			addr: Address{Scheme: SchemeHTTP, Address: "localhost:8080"},
			want: "http://localhost:8080",
		},
		{
			name: "HTTPS address",
			addr: Address{Scheme: SchemeHTTPS, Address: "example.com:443"},
			want: "https://example.com:443",
		},
		{
			name: "GRPC address",
			addr: Address{Scheme: SchemeGRPC, Address: "localhost:50051"},
			want: "grpc://localhost:50051",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.addr.String())
		})
	}
}
