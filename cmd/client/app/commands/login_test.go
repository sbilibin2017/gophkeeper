package commands

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sbilibin2017/gophkeeper/internal/models"
)

func TestParseLoginFlags(t *testing.T) {
	tests := []struct {
		name        string
		flags       map[string]string
		wantURL     string
		wantInt     bool
		expectError bool
	}{
		{"empty flags", map[string]string{}, "", false, false},
		{"only server-url", map[string]string{"server-url": "http://example.com"}, "http://example.com", false, false},
		{"interactive true", map[string]string{"interactive": "true"}, "", true, false},
		{"interactive false", map[string]string{"interactive": "false"}, "", false, false},
		{"invalid interactive", map[string]string{"interactive": "bad"}, "", false, true},
		{"both flags set", map[string]string{"server-url": "https://srv", "interactive": "true"}, "https://srv", true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotURL, gotInt, err := parseLoginFlags(tt.flags)

			if tt.expectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantURL, gotURL)
			assert.Equal(t, tt.wantInt, gotInt)
		})
	}
}

func TestParseLoginFlagsInteractive(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantUser  string
		wantPass  string
		wantError bool
	}{
		{"valid input", "john\nsecret\n", "john", "secret", false},
		{"empty input", "", "", "", true},
		{"only username", "john\n", "", "", true}, // добавил кейс, когда нет пароля
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.input)
			secret, err := parseLoginFlagsInteractive(reader)

			if tt.wantError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.wantUser, secret.Username)
			require.Equal(t, tt.wantPass, secret.Password)
		})
	}
}

func TestParseLoginArgs(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		wantUser  string
		wantPass  string
		wantError bool
	}{
		{"valid args", []string{"user", "pass"}, "user", "pass", false},
		{"missing password", []string{"user"}, "", "", true},
		{"no args", []string{}, "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			secret, err := parseLoginArgs(tt.args)

			if tt.wantError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.wantUser, secret.Username)
			require.Equal(t, tt.wantPass, secret.Password)
		})
	}
}

func TestValidateLoginRequest(t *testing.T) {
	tests := []struct {
		name      string
		secret    *models.UsernamePassword
		wantError bool
	}{
		{"nil secret", nil, true},
		{"empty username", &models.UsernamePassword{Username: "", Password: "pass"}, true},
		{"empty password", &models.UsernamePassword{Username: "user", Password: ""}, true},
		{"valid", &models.UsernamePassword{Username: "user", Password: "pass"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateLoginRequest(tt.secret)
			if tt.wantError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestNewLoginConfig(t *testing.T) {
	tests := []struct {
		name      string
		url       string
		wantError bool
	}{
		{"invalid prefix", "ftp://bad", true},
		{"http", "http://localhost", false},
		{"https", "https://localhost", false},
		{"grpc", "grpc://localhost", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := newLoginConfig(tt.url)
			if tt.wantError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestSetLoginEnv(t *testing.T) {
	tests := []struct {
		name      string
		serverURL string
		token     string
	}{
		{"with token", "http://localhost", "token-123"},
		{"empty token", "http://localhost", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := setLoginEnv(tt.serverURL, tt.token)
			require.NoError(t, err)

			assert.Equal(t, tt.serverURL, os.Getenv("GOPHKEEPER_SERVER_URL"))
			assert.Equal(t, tt.token, os.Getenv("GOPHKEEPER_TOKEN"))
		})
	}
}
