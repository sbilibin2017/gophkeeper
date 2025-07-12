package commands

import (
	"bufio"
	"context"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// cleanupDB удаляет файл базы данных после теста
func cleanupDB() {
	_ = os.Remove("gophkeeper.db")
}

func TestAddBinary_ValidFile(t *testing.T) {
	defer cleanupDB()

	tmpFile, err := os.CreateTemp("", "testbin")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	content := []byte("binary content test")
	_, err = tmpFile.Write(content)
	require.NoError(t, err)
	tmpFile.Close()

	err = addBinary(context.Background(), []string{tmpFile.Name()}, nil)
	require.NoError(t, err)
}

func TestAddBinary_FileNotExist(t *testing.T) {
	defer cleanupDB()

	err := addBinary(context.Background(), []string{"nonexistent.file"}, nil)
	require.Error(t, err)
}

func TestAddUsernamePassword_Args(t *testing.T) {
	defer cleanupDB()

	args := []string{"user", "pass"}
	meta := []string{"env=prod"}

	err := addUsernamePassword(context.Background(), args, false, meta, nil)
	require.NoError(t, err)
}

func TestAddUsernamePassword_Args_Missing(t *testing.T) {
	defer cleanupDB()

	args := []string{"onlyuser"}
	meta := []string{}

	err := addUsernamePassword(context.Background(), args, false, meta, nil)
	require.Error(t, err)
}

func TestAddUsernamePassword_Interactive(t *testing.T) {
	defer cleanupDB()

	input := `badMetaLine
anotherBadLine

testuser
testpass
`
	reader := bufio.NewReader(strings.NewReader(input))

	err := addUsernamePassword(context.Background(), []string{}, true, []string{}, reader)
	require.NoError(t, err)
}

func TestAddUsernamePassword_Interactive_MissingInput(t *testing.T) {
	defer cleanupDB()

	input := "" // пустой ввод, чтение даст ошибку EOF
	reader := bufio.NewReader(strings.NewReader(input))

	err := addUsernamePassword(context.Background(), []string{}, true, []string{}, reader)
	require.Error(t, err)
}

func TestAddText_Args(t *testing.T) {
	defer cleanupDB()

	args := []string{"Hello world"}
	meta := []string{"source=manual"}

	err := addText(context.Background(), args, false, meta, nil)
	require.NoError(t, err)
}

func TestAddText_Args_Missing(t *testing.T) {
	defer cleanupDB()

	args := []string{}
	meta := []string{}

	err := addText(context.Background(), args, false, meta, nil)
	require.Error(t, err)
}

func TestAddText_Interactive(t *testing.T) {
	defer cleanupDB()

	// Ввод: одна строка метаданных, затем текст в две строки, затем пустая строка для завершения текста
	input := "key=value\n\nline1\nline2\n\n"
	reader := bufio.NewReader(strings.NewReader(input))

	err := addText(context.Background(), []string{}, true, []string{}, reader)
	require.NoError(t, err)
}

func TestAddText_Interactive_EmptyInput(t *testing.T) {
	defer cleanupDB()

	// Ввод: пустая строка для метаданных, затем пустая строка для текста
	input := "\n\n"
	reader := bufio.NewReader(strings.NewReader(input))

	err := addText(context.Background(), []string{}, true, []string{}, reader)
	require.NoError(t, err) // пустой текст - не ошибка
}

func TestAddBankCard_Interactive(t *testing.T) {
	defer cleanupDB()

	// Ввод: метаданные, затем номер карты, срок действия, CVV
	input := `bank=testbank
owner=ivan ivanov

4111111111111111
12/30
321
`
	reader := bufio.NewReader(strings.NewReader(input))

	err := addBankCard(context.Background(), true, []string{}, reader)
	require.NoError(t, err)
}

func TestAddBankCard_Interactive_MissingFields(t *testing.T) {
	defer cleanupDB()

	// Пропущен CVV
	input := `type=test

4111111111111111
12/30
` // нет строки с CVV
	reader := bufio.NewReader(strings.NewReader(input))

	err := addBankCard(context.Background(), true, []string{}, reader)
	require.Error(t, err)
}

func TestAddBankCard_NonInteractive(t *testing.T) {
	defer cleanupDB()

	reader := bufio.NewReader(strings.NewReader("")) // пусто, не важно

	err := addBankCard(context.Background(), false, []string{}, reader)
	require.Error(t, err)
	require.Contains(t, err.Error(), "требуется использовать --interactive")
}
