package commands

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sbilibin2017/gophkeeper/internal/models"
)

// Тесты для parseRegisterFlags, проверяем парсинг флагов из map[string]string
func TestParseRegisterFlags(t *testing.T) {
	tests := []struct {
		name        string
		flags       map[string]string
		wantURL     string
		wantInt     bool
		expectError bool
	}{
		{
			name:    "empty flags",
			flags:   map[string]string{},
			wantURL: "",
			wantInt: false,
		},
		{
			name:    "only server-url set",
			flags:   map[string]string{"server-url": "http://example.com"},
			wantURL: "http://example.com",
			wantInt: false,
		},
		{
			name:    "interactive true",
			flags:   map[string]string{"interactive": "true"},
			wantURL: "",
			wantInt: true,
		},
		{
			name:    "interactive false",
			flags:   map[string]string{"interactive": "false"},
			wantURL: "",
			wantInt: false,
		},
		{
			name:        "invalid interactive",
			flags:       map[string]string{"interactive": "notabool"},
			expectError: true,
		},
		{
			name:    "both flags set",
			flags:   map[string]string{"server-url": "https://srv", "interactive": "true"},
			wantURL: "https://srv",
			wantInt: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotURL, gotInt, err := parseRegisterFlags(tt.flags)

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

// Тесты для parseRegisterFlagsInteractive, читаем username и password из io.Reader
func TestParseRegisterFlagsInteractive(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantUser  string
		wantPass  string
		wantError bool
	}{
		{"valid input", "testuser\nmypassword\n", "testuser", "mypassword", false},
		{"empty input", "", "", "", true},
		{"only username", "onlyuser\n", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.input)
			secret, err := parseRegisterFlagsInteractive(reader)

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

// Тесты для parseRegisterArgs — проверяем наличие username и password в args
func TestParseRegisterArgs(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		wantUser  string
		wantPass  string
		wantError bool
	}{
		{"valid args", []string{"user", "pass"}, "user", "pass", false},
		{"missing password", []string{"onlyone"}, "", "", true},
		{"empty args", []string{}, "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			secret, err := parseRegisterArgs(tt.args)

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

// Тесты для validateRegisterRequest — проверяем наличие обязательных полей
func TestValidateRegisterRequest(t *testing.T) {
	tests := []struct {
		name      string
		secret    *models.UsernamePassword
		wantError bool
	}{
		{"nil secret", nil, true},
		{"empty username", &models.UsernamePassword{Username: "", Password: "pass"}, true},
		{"empty password", &models.UsernamePassword{Username: "user", Password: ""}, true},
		{"valid secret", &models.UsernamePassword{Username: "user", Password: "pass"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateRegisterRequest(tt.secret)
			if tt.wantError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// Тесты для newRegisterConfig — проверка url на поддерживаемые протоколы
func TestNewRegisterConfig(t *testing.T) {
	tests := []struct {
		name      string
		url       string
		wantError bool
	}{
		{"unsupported protocol", "ftp://wrongprefix", true},
		{"http protocol", "http://localhost", false},
		{"https protocol", "https://localhost", false},
		{"grpc protocol", "grpc://localhost", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := newRegisterConfig(tt.url)
			if tt.wantError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// Тесты для setRegisterEnv — проверяем корректность установки env переменных
func TestSetRegisterEnv(t *testing.T) {
	tests := []struct {
		name      string
		serverURL string
		token     string
	}{
		{"set env vars", "http://localhost", "token123"},
		{"empty token", "http://localhost", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := setRegisterEnv(tt.serverURL, tt.token)
			require.NoError(t, err)

			assert.Equal(t, tt.serverURL, os.Getenv("GOPHKEEPER_SERVER_URL"))
			assert.Equal(t, tt.token, os.Getenv("GOPHKEEPER_TOKEN"))
		})
	}
}
