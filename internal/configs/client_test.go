package configs

import (
	"encoding/json"
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
	assert.Nil(t, cfg.GRPCClient)
}

func TestWithClientConfigHTTPClient(t *testing.T) {
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
			cfg, err := NewClientConfig(WithClientConfigHTTPClient(tt.baseURL))
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

func TestWithClientConfigDB(t *testing.T) {
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
			err := WithClientConfigDB(tt.dsns...)(cfg)
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

func TestWithClientConfigGRPCClient(t *testing.T) {
	tests := []struct {
		name       string
		addrs      []string
		wantErr    bool
		wantClient bool
	}{
		{
			name:       "empty address",
			addrs:      []string{""},
			wantErr:    false,
			wantClient: false,
		},
		{
			name:       "invalid address",
			addrs:      []string{"%$@#@!"},
			wantErr:    true,
			wantClient: false,
		},
		{
			name:       "valid address with no server",
			addrs:      []string{"localhost:50051"},
			wantErr:    false,
			wantClient: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &ClientConfig{}
			err := WithClientConfigGRPCClient(tt.addrs...)(cfg)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			if tt.wantClient {
				assert.NotNil(t, cfg.GRPCClient)
			} else {
				assert.Nil(t, cfg.GRPCClient)
			}
		})
	}
}

func TestPrepareMetaJSON(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantNil    bool
		wantString string
		wantErr    bool
	}{
		{
			name:    "empty string returns nil",
			input:   "",
			wantNil: true,
			wantErr: false,
		},
		{
			name:       "valid JSON returns normalized string",
			input:      `{"key1":"value1","key2":"value2"}`,
			wantNil:    false,
			wantString: `{"key1":"value1","key2":"value2"}`,
			wantErr:    false,
		},
		{
			name:    "invalid JSON returns error",
			input:   `{"key1":value1"}`,
			wantNil: true,
			wantErr: true,
		},
		{
			name:       "valid JSON with unordered keys",
			input:      `{"key2":"value2","key1":"value1"}`,
			wantNil:    false,
			wantString: `{"key1":"value1","key2":"value2"}`,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := PrepareMetaJSON(tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
				return
			}

			assert.NoError(t, err)

			if tt.wantNil {
				assert.Nil(t, got)
			} else {
				assert.NotNil(t, got)

				var gotMap, wantMap map[string]string
				err1 := json.Unmarshal([]byte(*got), &gotMap)
				err2 := json.Unmarshal([]byte(tt.wantString), &wantMap)
				assert.NoError(t, err1)
				assert.NoError(t, err2)
				assert.Equal(t, wantMap, gotMap)
			}
		})
	}
}
