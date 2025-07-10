package app

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewConfigCommand_RunE(t *testing.T) {
	cmd := newConfigCommand()

	// Ошибка, если ни один флаг не указан
	err := cmd.RunE(cmd, []string{})
	require.Error(t, err)
	require.Contains(t, err.Error(), "необходимо указать хотя бы один флаг")

	// Указываем только токен - ошибок не должно быть
	err = cmd.Flags().Set("token", "mytoken123")
	require.NoError(t, err)

	err = cmd.RunE(cmd, []string{})
	require.NoError(t, err)

	// Сбрасываем флаг токена для чистоты теста
	_ = cmd.Flags().Set("token", "")

	// Указываем только сервер - ошибок не должно быть
	err = cmd.Flags().Set("server-url", "https://example.com")
	require.NoError(t, err)

	err = cmd.RunE(cmd, []string{})
	require.NoError(t, err)

	// Сбрасываем флаг сервера для чистоты теста
	_ = cmd.Flags().Set("server-url", "")

	// Указываем и токен и сервер - ошибок не должно быть
	err = cmd.Flags().Set("token", "token123")
	require.NoError(t, err)
	err = cmd.Flags().Set("server-url", "https://server.com")
	require.NoError(t, err)

	err = cmd.RunE(cmd, []string{})
	require.NoError(t, err)
}
