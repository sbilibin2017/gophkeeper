package app

import (
	"bufio"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseListFlags_NonInteractive_Success(t *testing.T) {
	token := "test-token"
	serverURL := "https://server.com"
	interactive := false

	err := parseListFlags(&token, &serverURL, &interactive)
	require.NoError(t, err)
	assert.Equal(t, "test-token", token)
	assert.Equal(t, "https://server.com", serverURL)
}

func TestParseListFlags_NonInteractive_MissingToken(t *testing.T) {
	token := ""
	serverURL := "https://server.com"
	interactive := false

	err := parseListFlags(&token, &serverURL, &interactive)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "необходимо указать токен и URL сервера")
}

func TestParseListFlags_NonInteractive_MissingServerURL(t *testing.T) {
	token := "token"
	serverURL := ""
	interactive := false

	err := parseListFlags(&token, &serverURL, &interactive)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "необходимо указать токен и URL сервера")
}

func TestNewListCommand_Flags(t *testing.T) {
	cmd := newListCommand()

	err := cmd.Flags().Set("type", "login")
	require.NoError(t, err)
	assert.Equal(t, "login", secretType)

	err = cmd.Flags().Set("output-type", "file")
	require.NoError(t, err)
	assert.Equal(t, "file", outputType)

	err = cmd.Flags().Set("token", "token123")
	require.NoError(t, err)
	err = cmd.Flags().Set("server-url", "https://server.com")
	require.NoError(t, err)

	interactive := false
	token := "token123"
	serverURL := "https://server.com"

	err = parseListFlags(&token, &serverURL, &interactive)
	require.NoError(t, err)
}

func TestParseListFlagsInteractive(t *testing.T) {
	input := strings.Join([]string{
		"mytokenlist",         // токен
		"https://list.server", // URL сервера
	}, "\n") + "\n" // важен финальный перевод строки

	reader := bufio.NewReader(strings.NewReader(input))

	var token, serverURL string

	err := parseListFlagsInteractive(reader, &token, &serverURL)
	require.NoError(t, err)

	assert.Equal(t, "mytokenlist", token)
	assert.Equal(t, "https://list.server", serverURL)
}
