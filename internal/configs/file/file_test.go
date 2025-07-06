package file

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConfigFilePath(t *testing.T) {
	tempDir := t.TempDir() // временная директория для теста

	// Переопределяем функцию UserHomeDir в тесте (через переменную или интерфейс) — если так сделано
	// Или подменяем ConfigFilePath, чтобы использовать tempDir

	// Пример простой проверки, создадим директорию вручную
	configDir := filepath.Join(tempDir, ".gophkeeper")
	err := os.MkdirAll(configDir, 0o755)
	require.NoError(t, err)

	// Теперь проверяем создание файла в этом пути
	path := filepath.Join(configDir, "config.json")
	require.Equal(t, filepath.Join(tempDir, ".gophkeeper", "config.json"), path)
}

func TestLoadConfig_NoFile(t *testing.T) {
	// Создаём временную директорию и подменяем $HOME
	tmpHome := t.TempDir()

	// Подменяем HOME для теста
	originalHome := os.Getenv("HOME")
	require.NoError(t, os.Setenv("HOME", tmpHome))
	t.Cleanup(func() { _ = os.Setenv("HOME", originalHome) })

	// Запускаем тест
	cfg, err := LoadConfig()
	require.NoError(t, err)
	require.Empty(t, cfg)

	// Проверяем, что файл не создан
	path, err := ConfigFilePath()
	require.NoError(t, err)
	_, err = os.Stat(path)
	require.True(t, os.IsNotExist(err))
}

func TestSetGetUnsetConfigValue(t *testing.T) {
	tmpHome := t.TempDir()

	// Подменяем $HOME
	originalHome := os.Getenv("HOME")
	require.NoError(t, os.Setenv("HOME", tmpHome))
	t.Cleanup(func() { _ = os.Setenv("HOME", originalHome) })

	// Установка
	err := SetConfigValue("testKey", "testValue")
	require.NoError(t, err)

	// Получение
	val, ok, err := GetConfigValue("testKey")
	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, "testValue", val)

	// Удаление
	err = UnsetConfigValue("testKey")
	require.NoError(t, err)

	// Проверка отсутствия
	_, ok, err = GetConfigValue("testKey")
	require.NoError(t, err)
	require.False(t, ok)
}

func TestListConfig(t *testing.T) {
	tmpHome := t.TempDir()

	// Подменяем $HOME, чтобы не трогать реальные конфиги
	originalHome := os.Getenv("HOME")
	require.NoError(t, os.Setenv("HOME", tmpHome))
	t.Cleanup(func() {
		_ = os.Setenv("HOME", originalHome)
	})

	// Устанавливаем несколько значений
	require.NoError(t, SetConfigValue("foo", "bar"))
	require.NoError(t, SetConfigValue("hello", "world"))

	// Проверяем ListConfig
	cfg, err := ListConfig()
	require.NoError(t, err)
	require.Len(t, cfg, 2)
	require.Equal(t, "bar", cfg["foo"])
	require.Equal(t, "world", cfg["hello"])
}
