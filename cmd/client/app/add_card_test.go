package app

import (
	"bufio"
	"strings"
	"testing"

	"github.com/sbilibin2017/gophkeeper/cmd/client/app/flags"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseCardFlags_NonInteractive(t *testing.T) {
	cardNumber = "4111111111111111"
	cardExp = "12/25"
	cardCVV = "123"
	cardMeta = make(flags.MetaFlag) // очистим метаданные

	token := "token"
	serverURL := "https://server.com"
	interactive := false

	err := parseCardFlags(&token, &serverURL, &interactive)
	require.NoError(t, err)

	assert.Equal(t, "4111111111111111", cardNumber)
	assert.Equal(t, "12/25", cardExp)
	assert.Equal(t, "123", cardCVV)
	assert.Equal(t, "token", token)
	assert.Equal(t, "https://server.com", serverURL)
}

func TestParseCardFlags_MissingRequiredFields(t *testing.T) {
	cardNumber = ""
	cardExp = "12/25"
	cardCVV = "123"
	cardMeta = make(flags.MetaFlag)

	token := "token"
	serverURL := "https://server.com"
	interactive := false

	err := parseCardFlags(&token, &serverURL, &interactive)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "параметры number, expiry и cvv обязательны для заполнения")
}

func TestParseCardFlags_MissingTokenOrServerURL(t *testing.T) {
	cardNumber = "4111111111111111"
	cardExp = "12/25"
	cardCVV = "123"
	cardMeta = make(flags.MetaFlag)

	token := ""
	serverURL := ""
	interactive := false

	err := parseCardFlags(&token, &serverURL, &interactive)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "токен и URL сервера должны быть заданы")
}

func TestNewAddCardCommand_RunE_NonInteractive(t *testing.T) {
	cardMeta = make(flags.MetaFlag) // очистим метаданные

	cmd := newAddCardCommand()

	err := cmd.Flags().Set("number", "4111111111111111")
	require.NoError(t, err)
	err = cmd.Flags().Set("expiry", "12/25")
	require.NoError(t, err)
	err = cmd.Flags().Set("cvv", "123")
	require.NoError(t, err)
	err = cmd.Flags().Set("token", "test-token")
	require.NoError(t, err)
	err = cmd.Flags().Set("server-url", "https://server.com")
	require.NoError(t, err)
	err = cmd.Flags().Set("meta", "owner=John")
	require.NoError(t, err)

	// Синхронизируем глобальные переменные с флагами, чтобы RunE их увидел
	cardNumber, _ = cmd.Flags().GetString("number")
	cardExp, _ = cmd.Flags().GetString("expiry")
	cardCVV, _ = cmd.Flags().GetString("cvv")

	err = cmd.RunE(cmd, []string{})
	require.NoError(t, err)
}

func TestParseCardFlagsInteractive(t *testing.T) {
	// Входные данные для интерактивного ввода
	input := strings.Join([]string{
		"1234 5678 9012 3456", // номер карты
		"12/34",               // срок действия
		"123",                 // CVV
		"key1=value1",         // метаданные
		"key2=value2",
		"",                   // пустая строка для завершения ввода метаданных
		"mytoken123",         // токен
		"https://server.url", // URL сервера
	}, "\n") + "\n" // ВАЖНО: добавить финальный перевод строки, чтобы ReadString не получил EOF

	reader := bufio.NewReader(strings.NewReader(input))

	var token, serverURL string

	// Обнуляем глобальные переменные, если нужно
	cardNumber = ""
	cardExp = ""
	cardCVV = ""
	cardMeta = make(flags.MetaFlag)

	err := parseCardFlagsInteractive(reader, &token, &serverURL)
	require.NoError(t, err)

	assert.Equal(t, "1234 5678 9012 3456", cardNumber)
	assert.Equal(t, "12/34", cardExp)
	assert.Equal(t, "123", cardCVV)

	expectedMeta := flags.MetaFlag{
		"key1": "value1",
		"key2": "value2",
	}
	assert.Equal(t, expectedMeta, cardMeta)

	assert.Equal(t, "mytoken123", token)
	assert.Equal(t, "https://server.url", serverURL)
}
