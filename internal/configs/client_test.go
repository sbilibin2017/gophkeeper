package configs

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClientConfig(t *testing.T) {
	cfg, err := NewClientConfig()
	require.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Nil(t, cfg.HTTPClient)
	assert.Nil(t, cfg.DB)
}

func TestWithHTTPClient(t *testing.T) {
	tests := []struct {
		name    string
		baseURL string
		wantNil bool
	}{
		{"empty baseURL", "", true},
		{"valid baseURL", "http://localhost:8080", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := NewClientConfig(WithHTTPClient(tt.baseURL))
			require.NoError(t, err)
			if tt.wantNil {
				assert.Nil(t, cfg.HTTPClient)
			} else {
				require.NotNil(t, cfg.HTTPClient)
				assert.Equal(t, tt.baseURL, cfg.HTTPClient.BaseURL)
			}
		})
	}
}

func TestWithDB(t *testing.T) {
	tests := []struct {
		name      string
		dsns      []string
		wantError bool
		wantDBNil bool
	}{
		{"empty DSN", []string{""}, false, true},
		{"invalid DSN", []string{"/invalid/path/to/db"}, true, true},
		{"valid DSN", []string{":memory:"}, false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &ClientConfig{}
			err := WithDB(tt.dsns...)(cfg)
			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			if tt.wantDBNil {
				assert.Nil(t, cfg.DB)
			} else {
				assert.NotNil(t, cfg.DB)
			}
		})
	}
}
