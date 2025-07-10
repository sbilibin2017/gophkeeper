package app

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewVersionCommand_Output(t *testing.T) {
	// Подготовка буфера для захвата вывода команды
	buf := new(bytes.Buffer)

	// Создание команды version
	cmd := newVersionCommand()
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	// Выполнение команды
	err := cmd.Execute()
	assert.NoError(t, err)

	// Получение и проверка вывода
	output := buf.String()
	assert.Contains(t, output, "version:")
	assert.Contains(t, output, "date:")
	assert.True(t, strings.Contains(output, buildVersion), "вывод должен содержать buildVersion")
	assert.True(t, strings.Contains(output, buildDate), "вывод должен содержать buildDate")
}
