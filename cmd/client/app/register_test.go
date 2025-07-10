package app

import (
	"bufio"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRegisterCommand_Flags(t *testing.T) {
	cmd := newRegisterCommand()

	if cmd == nil {
		t.Fatal("expected command, got nil")
	}
	if cmd.Use != "register" {
		t.Errorf("expected Use='register', got '%s'", cmd.Use)
	}

	flags := cmd.Flags()

	if flags.Lookup("username") == nil {
		t.Error("username flag is not registered")
	}
	if flags.Lookup("password") == nil {
		t.Error("password flag is not registered")
	}
	if flags.Lookup("server-url") == nil {
		t.Error("server-url flag is not registered")
	}
	if flags.Lookup("interactive") == nil {
		t.Error("interactive flag is not registered")
	}
}

func TestParseRegisterFlagsInteractive(t *testing.T) {
	input := strings.Join([]string{
		"newuser",            // username
		"newpassword",        // password
		"https://reg.server", // server URL
	}, "\n") + "\n" // ВАЖНО: финальный перевод строки для корректного чтения

	reader := bufio.NewReader(strings.NewReader(input))

	var serverURL string

	// Обнуляем глобальные переменные перед тестом
	registerUsername = ""
	registerPassword = ""

	err := parseRegisterFlagsInteractive(reader, &serverURL)
	require.NoError(t, err)

	assert.Equal(t, "newuser", registerUsername)
	assert.Equal(t, "newpassword", registerPassword)
	assert.Equal(t, "https://reg.server", serverURL)
}

func TestParseRegisterFlags_NonInteractive_Valid(t *testing.T) {
	registerUsername = "reguser"
	registerPassword = "regpass"

	serverURL := "https://server2"
	interactive := false

	err := parseRegisterFlags(&serverURL, &interactive)
	require.NoError(t, err)

	assert.Equal(t, "reguser", registerUsername)
	assert.Equal(t, "regpass", registerPassword)
	assert.Equal(t, "https://server2", serverURL)
}

func TestParseRegisterFlags_NonInteractive_MissingUserOrPass(t *testing.T) {
	registerUsername = ""
	registerPassword = ""

	serverURL := "https://server2"
	interactive := false

	err := parseRegisterFlags(&serverURL, &interactive)
	require.Error(t, err)

}
