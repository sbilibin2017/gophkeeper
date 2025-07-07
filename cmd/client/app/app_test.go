package app

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

func TestNewAppCommand(t *testing.T) {
	cmd := NewAppCommand()
	require.NotNil(t, cmd)

	// Проверяем основные поля команды
	require.Equal(t, "gophkeeper", cmd.Use)
	require.Contains(t, cmd.Short, "CLI-инструмент")
	require.Contains(t, cmd.Long, "Доступные команды:")

	// Проверяем, что добавлены дочерние команды
	subCmds := cmd.Commands()
	require.NotEmpty(t, subCmds)

	// Список имен дочерних команд, которые мы ожидаем
	expectedCmds := []string{
		"build-info",
		"register",
		"login",
		"add",
		"get",
		"list",
		"sync",
	}

	// Проверяем, что все ожидаемые команды есть
	for _, name := range expectedCmds {
		require.Truef(t, containsCommand(subCmds, name), "ожидается команда %q", name)
	}
}

// Вспомогательная функция проверяет, есть ли команда с таким именем в срезе
func containsCommand(cmds []*cobra.Command, name string) bool {
	for _, c := range cmds {
		if c.Name() == name {
			return true
		}
	}
	return false
}
