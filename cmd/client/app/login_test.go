package app

import (
	"bufio"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseLoginFlagsInteractive(t *testing.T) {
	input := strings.Join([]string{
		"testuser",             // имя пользователя
		"supersecret",          // пароль
		"https://login.server", // URL сервера
	}, "\n") + "\n" // обязательно финальный перевод строки

	reader := bufio.NewReader(strings.NewReader(input))

	var serverURL string

	// Обнуляем глобальные переменные перед тестом
	loginUsername = ""
	loginPassword = ""

	err := parseLoginFlagsInteractive(reader, &serverURL)
	require.NoError(t, err)

	assert.Equal(t, "testuser", loginUsername)
	assert.Equal(t, "supersecret", loginPassword)
	assert.Equal(t, "https://login.server", serverURL)
}

func TestParseLoginFlags_NonInteractive_Valid(t *testing.T) {
	loginUsername = "user1"
	loginPassword = "pass1"

	serverURL := "https://server1"
	interactive := false

	err := parseLoginFlags(&serverURL, &interactive)
	require.NoError(t, err)
	assert.Equal(t, "user1", loginUsername)
	assert.Equal(t, "pass1", loginPassword)
	assert.Equal(t, "https://server1", serverURL)
}

func TestParseLoginFlags_NonInteractive_MissingUserOrPass(t *testing.T) {
	loginUsername = ""
	loginPassword = ""
	serverURL := "https://server1"
	interactive := false

	err := parseLoginFlags(&serverURL, &interactive)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "имя пользователя и пароль не могут быть пустыми")
}
