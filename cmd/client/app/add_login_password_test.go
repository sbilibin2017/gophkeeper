package app

import (
	"bufio"
	"strings"
	"testing"

	"github.com/sbilibin2017/gophkeeper/cmd/client/app/flags"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAddLoginPasswordCommand_RunE_NonInteractive(t *testing.T) {
	// Очистим глобальные переменные
	loginPasswordUsername = ""
	loginPasswordPassword = ""
	loginPasswordMeta = make(flags.MetaFlag)

	cmd := newAddLoginPasswordCommand()

	err := cmd.Flags().Set("username", "user123")
	require.NoError(t, err)
	err = cmd.Flags().Set("password", "secret")
	require.NoError(t, err)
	err = cmd.Flags().Set("token", "test-token")
	require.NoError(t, err)
	err = cmd.Flags().Set("server-url", "https://server.com")
	require.NoError(t, err)
	err = cmd.Flags().Set("meta", "site=example.com")
	require.NoError(t, err)

	// Синхронизируем глобальные переменные с флагами, чтобы RunE их увидел
	loginPasswordUsername, _ = cmd.Flags().GetString("username")
	loginPasswordPassword, _ = cmd.Flags().GetString("password")

	err = cmd.RunE(cmd, []string{})
	require.NoError(t, err)
	require.Equal(t, "user123", loginPasswordUsername)
	require.Equal(t, "secret", loginPasswordPassword)
	require.Contains(t, loginPasswordMeta, "site")
	require.Equal(t, "example.com", loginPasswordMeta["site"])
}

func TestParseLoginPasswordFlags_MissingRequired(t *testing.T) {
	// Тест с отсутствующими обязательными параметрами
	loginPasswordUsername = ""
	loginPasswordPassword = ""
	token := "token"
	serverURL := "https://server"
	interactive := false

	err := parseLoginPasswordFlags(&token, &serverURL, &interactive)
	require.Error(t, err)
	require.Contains(t, err.Error(), "параметры username и password обязательны")
}

func TestParseLoginPasswordFlags_MissingTokenOrURL(t *testing.T) {
	loginPasswordUsername = "user"
	loginPasswordPassword = "pass"
	token := ""
	serverURL := ""
	interactive := false

	err := parseLoginPasswordFlags(&token, &serverURL, &interactive)
	require.Error(t, err)
	require.Contains(t, err.Error(), "токен и URL сервера должны быть заданы")
}

func TestParseLoginPasswordFlagsInteractive(t *testing.T) {
	input := strings.Join([]string{
		"testuser",       // логин
		"secretpass",     // пароль
		"env=production", // метаданные
		"version=1.0",
		"",                     // пустая строка для завершения ввода метаданных
		"mytokenlogin",         // токен
		"https://login.server", // URL сервера
	}, "\n") + "\n" // финальный перевод строки чтобы ReadString не получил EOF

	reader := bufio.NewReader(strings.NewReader(input))

	var token, serverURL string

	// Обнуляем глобальные переменные
	loginPasswordUsername = ""
	loginPasswordPassword = ""
	loginPasswordMeta = make(flags.MetaFlag)

	err := parseLoginPasswordFlagsInteractive(reader, &token, &serverURL)
	require.NoError(t, err)

	assert.Equal(t, "testuser", loginPasswordUsername)
	assert.Equal(t, "secretpass", loginPasswordPassword)

	expectedMeta := flags.MetaFlag{
		"env":     "production",
		"version": "1.0",
	}
	assert.Equal(t, expectedMeta, loginPasswordMeta)

	assert.Equal(t, "mytokenlogin", token)
	assert.Equal(t, "https://login.server", serverURL)
}
