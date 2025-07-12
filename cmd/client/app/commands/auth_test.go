package commands

import (
	"bufio"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sbilibin2017/gophkeeper/internal/models"
)

// Тесты для parseAuthFlags — проверка парсинга флагов из map[string]string
func TestParseAuthFlags(t *testing.T) {
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
			gotURL, gotInt, err := parseAuthFlags(tt.flags)

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

// Тесты для parseAuthFlagsInteractive — читаем username, password и meta из bufio.Reader
func TestParseAuthFlagsInteractive(t *testing.T) {
	// Заготовка с двумя строками: логин, пароль и пустые метаданные (имитируем)
	input := "john\nsecret\n\n"

	reader := bufio.NewReader(strings.NewReader(input))
	secret, err := parseAuthFlagsInteractive(reader)
	require.NoError(t, err)
	assert.Equal(t, "john", secret.Username)
	assert.Equal(t, "secret", secret.Password)
	// Meta зависит от реализации parsemeta.ParseMetaInteractive, можно проверить, что не nil
	assert.NotNil(t, secret.Meta)
}

// Тесты для parseAuthArgs — проверка аргументов
func TestParseAuthArgs(t *testing.T) {
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
			secret, err := parseAuthArgs(tt.args)

			if tt.wantError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantUser, secret.Username)
			assert.Equal(t, tt.wantPass, secret.Password)
		})
	}
}

// Тесты для validateAuthRequest — проверка обязательных полей
func TestValidateAuthRequest(t *testing.T) {
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
			err := validateAuthRequest(tt.secret)
			if tt.wantError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// Тесты для setAuthEnv — установка env переменных
func TestSetAuthEnv(t *testing.T) {
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
			err := setAuthEnv(tt.serverURL, tt.token)
			require.NoError(t, err)

			assert.Equal(t, tt.serverURL, os.Getenv("GOPHKEEPER_SERVER_URL"))
			assert.Equal(t, tt.token, os.Getenv("GOPHKEEPER_TOKEN"))
		})
	}
}
