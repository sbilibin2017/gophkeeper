package app

import (
	"bufio"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseSyncFlags_NonInteractive_ValidInput(t *testing.T) {
	token := "mytoken"
	serverURL := "https://example.com"
	resolver := "server"
	interactive := false

	err := parseSyncFlags(&token, &serverURL, &resolver, &interactive)
	require.NoError(t, err)
	assert.Equal(t, "mytoken", token)
	assert.Equal(t, "https://example.com", serverURL)
	assert.Equal(t, "server", resolver)
}

func TestParseSyncFlags_MissingTokenAndServerURL(t *testing.T) {
	token := ""
	serverURL := ""
	resolver := ""
	interactive := false

	os.Unsetenv("GOPHKEEPER_TOKEN")
	os.Unsetenv("GOPHKEEPER_SERVER_URL")

	err := parseSyncFlags(&token, &serverURL, &resolver, &interactive)
	require.Error(t, err)

}

func TestParseSyncFlags_UsesEnvFallback(t *testing.T) {
	_ = os.Setenv("GOPHKEEPER_TOKEN", "envtoken")
	_ = os.Setenv("GOPHKEEPER_SERVER_URL", "https://envserver.com")
	defer os.Unsetenv("GOPHKEEPER_TOKEN")
	defer os.Unsetenv("GOPHKEEPER_SERVER_URL")

	token := ""
	serverURL := ""
	resolver := ""
	interactive := false

	err := parseSyncFlags(&token, &serverURL, &resolver, &interactive)
	require.NoError(t, err)
	assert.Equal(t, "envtoken", token)
	assert.Equal(t, "https://envserver.com", serverURL)
}

func TestParseSyncFlags_InvalidResolver(t *testing.T) {
	token := "t"
	serverURL := "url"
	resolver := "wrong"
	interactive := false

	err := parseSyncFlags(&token, &serverURL, &resolver, &interactive)
	require.Error(t, err)

}

func TestNewSyncCommand_RunE_ValidFlags(t *testing.T) {
	cmd := newSyncCommand()

	err := cmd.Flags().Set("token", "t")
	require.NoError(t, err)
	err = cmd.Flags().Set("server-url", "https://server")
	require.NoError(t, err)
	err = cmd.Flags().Set("resolver", "interactive")
	require.NoError(t, err)

	err = cmd.RunE(cmd, []string{})
	require.NoError(t, err)
}

func TestParseSyncFlagsInteractive(t *testing.T) {
	input := strings.Join([]string{
		"mytoken123",          // token
		"https://sync.server", // server URL
		"interactive",         // conflict resolution strategy
	}, "\n") + "\n" // важен финальный перевод строки

	reader := bufio.NewReader(strings.NewReader(input))

	var token, serverURL, resolver string

	err := parseSyncFlagsInteractive(reader, &token, &serverURL, &resolver)
	require.NoError(t, err)

	assert.Equal(t, "mytoken123", token)
	assert.Equal(t, "https://sync.server", serverURL)
	assert.Equal(t, "interactive", resolver)
}
