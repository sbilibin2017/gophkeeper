package app

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigureCommand_Success(t *testing.T) {
	// Создаем временный каталог
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".gophkeeper")
	t.Setenv("HOME", tmpDir)

	token := "test-jwt-token"
	cmd := newConfigureCommand()
	cmd.SetArgs([]string{"--token", token})

	err := cmd.Execute()
	require.NoError(t, err)

	// Проверяем, что файл записан
	data, err := os.ReadFile(configPath)
	require.NoError(t, err)
	assert.Equal(t, token, string(data))
}

func TestConfigureCommand_MissingToken(t *testing.T) {
	cmd := newConfigureCommand()
	cmd.SetArgs([]string{}) // без флага token

	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "token is required")
}

func TestConfigureCommand_FileWriteError(t *testing.T) {
	// Устанавливаем $HOME в несуществующий путь (или без прав доступа)
	t.Setenv("HOME", "/root/should-fail") // для Linux-систем

	cmd := newConfigureCommand()
	cmd.SetArgs([]string{"--token", "test-token"})

	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to save token")
}
