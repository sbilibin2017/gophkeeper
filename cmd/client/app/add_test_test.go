package app

import (
	"bufio"
	"strings"
	"testing"

	"github.com/sbilibin2017/gophkeeper/cmd/client/app/flags"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAddTextCommand_RunE_NonInteractive(t *testing.T) {
	// Очистка глобальных переменных перед тестом
	textData = ""
	textMeta = make(flags.MetaFlag)

	cmd := newAddTextCommand()

	err := cmd.Flags().Set("data", "секретные заметки")
	require.NoError(t, err)
	err = cmd.Flags().Set("token", "test-token")
	require.NoError(t, err)
	err = cmd.Flags().Set("server-url", "https://server.com")
	require.NoError(t, err)
	err = cmd.Flags().Set("meta", "note=личное")
	require.NoError(t, err)

	// Синхронизируем глобальные переменные с флагами
	textData, _ = cmd.Flags().GetString("data")

	err = cmd.RunE(cmd, []string{})
	require.NoError(t, err)
	require.Equal(t, "секретные заметки", textData)
	require.Contains(t, textMeta, "note")
	require.Equal(t, "личное", textMeta["note"])
}

func TestParseTextFlags_MissingData(t *testing.T) {
	textData = ""
	token := "token"
	serverURL := "https://server"
	interactive := false

	err := parseTextFlags(&token, &serverURL, &interactive)
	require.Error(t, err)
	require.Contains(t, err.Error(), "параметр data обязателен")
}

func TestParseTextFlags_MissingTokenOrURL(t *testing.T) {
	textData = "some data"
	token := ""
	serverURL := ""
	interactive := false

	err := parseTextFlags(&token, &serverURL, &interactive)
	require.Error(t, err)
	require.Contains(t, err.Error(), "токен и URL сервера должны быть заданы")
}

func TestParseTextFlagsInteractive(t *testing.T) {
	input := strings.Join([]string{
		"Пример текстовых данных", // текстовые данные
		"author=John Doe", // метаданные
		"lang=ru",
		"",                    // пустая строка для завершения ввода метаданных
		"mytokentext",         // токен
		"https://text.server", // URL сервера
	}, "\n") + "\n" // добавляем финальный перевод строки, чтобы ReadString не получил EOF

	reader := bufio.NewReader(strings.NewReader(input))

	var token, serverURL string

	// Обнуляем глобальные переменные
	textData = ""
	textMeta = make(flags.MetaFlag)

	err := parseTextFlagsInteractive(reader, &token, &serverURL)
	require.NoError(t, err)

	assert.Equal(t, "Пример текстовых данных", textData)

	expectedMeta := flags.MetaFlag{
		"author": "John Doe",
		"lang":   "ru",
	}
	assert.Equal(t, expectedMeta, textMeta)

	assert.Equal(t, "mytokentext", token)
	assert.Equal(t, "https://text.server", serverURL)
}
