package commands

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewCommand(t *testing.T) {
	cmd := NewRootCommand()
	require.NotNil(t, cmd, "корневая команда не должна быть nil")
	require.Equal(t, "gophkeeper-client", cmd.Use, "имя команды должно быть gophkeeper-client")

	// Проверяем вывод help
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"--help"})
	err := cmd.Execute()
	require.NoError(t, err)

}

func TestNewCommand_RunE(t *testing.T) {
	ctx := context.Background()

	mockDeviceIDGetter := func() (string, error) {
		return "device123", nil
	}

	mockHTTPRunner := func(
		ctx context.Context,
		serverURL string,
		databaseMigrationsDir string,
		username string,
		password string,
		deviceID string,
	) ([]byte, string, error) {
		require.Equal(t, "device123", deviceID)
		require.Equal(t, "http://localhost:8080", serverURL)
		require.Equal(t, "user1", username)
		require.Equal(t, "pass1", password)
		return []byte("mocked_priv_key"), "mocked_token", nil
	}

	cmd := NewRegisterCommand(mockHTTPRunner, mockDeviceIDGetter)

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	cmd.SetArgs([]string{
		"--username", "user1",
		"--password", "pass1",
	})

	err := cmd.ExecuteContext(ctx)
	require.NoError(t, err)

	output := buf.String()
	require.Contains(t, output, "Регистрация успешна")
	require.Contains(t, output, "Приватный ключ: mocked_priv_key")
	require.Contains(t, output, "Токен: mocked_token")
}

func TestNewCommand_DeviceIDGetterError(t *testing.T) {
	ctx := context.Background()

	mockDeviceIDGetter := func() (string, error) {
		return "", errors.New("device ID error")
	}

	mockHTTPRunner := func(
		ctx context.Context,
		serverURL string,
		databaseMigrationsDir string,
		username string,
		password string,
		deviceID string,
	) ([]byte, string, error) {
		return []byte("mocked_priv_key"), "mocked_token", nil
	}

	cmd := NewRegisterCommand(mockHTTPRunner, mockDeviceIDGetter)
	cmd.SetArgs([]string{
		"--username", "user1",
		"--password", "pass1",
	})

	err := cmd.ExecuteContext(ctx)
	require.Error(t, err)
	require.EqualError(t, err, "device ID error")
}
