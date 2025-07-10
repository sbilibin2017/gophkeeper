package app

import (
	"bufio"
	"strings"
	"testing"

	"github.com/sbilibin2017/gophkeeper/cmd/client/app/flags"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseBinaryFlags_NonInteractive(t *testing.T) {
	content = "file.bin"
	token := "token"
	serverURL := "https://server.com"
	interactive := false

	err := parseBinaryFlags(&token, &serverURL, &interactive)
	require.NoError(t, err)
	assert.Equal(t, "file.bin", content)
	assert.Equal(t, "token", token)
	assert.Equal(t, "https://server.com", serverURL)
}

func TestParseBinaryFlags_MissingContent(t *testing.T) {
	content = ""
	token := "token"
	serverURL := "https://server.com"
	interactive := false

	err := parseBinaryFlags(&token, &serverURL, &interactive)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "parameter content is required")
}

func TestParseBinaryFlags_MissingTokenOrServerURL(t *testing.T) {
	content = "file.bin"
	token := ""
	serverURL := ""
	interactive := false

	err := parseBinaryFlags(&token, &serverURL, &interactive)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "token and server URL must be provided")
}

func TestNewAddBinaryCommand_RunE_NonInteractive(t *testing.T) {
	content = "file.bin"
	meta = make(flags.MetaFlag) // очистим глобальную мета

	cmd := newAddBinaryCommand()

	err := cmd.Flags().Set("token", "test-token")
	require.NoError(t, err)
	err = cmd.Flags().Set("server-url", "https://server.com")
	require.NoError(t, err)
	err = cmd.Flags().Set("content", "file.bin")
	require.NoError(t, err)
	err = cmd.Flags().Set("meta", "site=example.com")
	require.NoError(t, err)

	err = cmd.RunE(cmd, []string{})
	require.NoError(t, err)
}

func TestParseBinaryFlagsInteractive_WithReader(t *testing.T) {
	input := `myfile.bin
site=example.com
env=prod

mytoken
https://server.com
`

	reader := bufio.NewReader(strings.NewReader(input))
	meta = make(flags.MetaFlag)
	content = ""

	var token, serverURL string
	err := parseBinaryFlagsInteractive(reader, &token, &serverURL)
	require.NoError(t, err)

	assert.Equal(t, "myfile.bin", content)
	assert.Equal(t, "mytoken", token)
	assert.Equal(t, "https://server.com", serverURL)
	assert.Equal(t, flags.MetaFlag{
		"site": "example.com",
		"env":  "prod",
	}, meta)
}
