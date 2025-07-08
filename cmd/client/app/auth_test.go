package app

import (
	"context"
	"testing"
	"time"

	"github.com/sbilibin2017/gophkeeper/internal/configs"
	"github.com/sbilibin2017/gophkeeper/internal/models"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Чтобы тесты не упирались в cobra.Command напрямую, создадим вспомогательные методы
func prepareCmdWithFlags(cmd *cobra.Command, flags map[string]string) error {
	for k, v := range flags {
		if err := cmd.Flags().Set(k, v); err != nil {
			return err
		}
	}
	return nil
}

func TestParseRegisterFlags_Flags(t *testing.T) {
	cmd := newRegisterCommand()
	err := prepareCmdWithFlags(cmd, map[string]string{
		"server-url": "grpc://localhost:8080",
		"username":   "flaguser",
		"password":   "flagpass",
	})
	require.NoError(t, err)

	config, creds, err := parseRegisterFlags(cmd)
	require.NoError(t, err)
	assert.Equal(t, "flaguser", creds.Username)
	assert.Equal(t, "flagpass", creds.Password)
	assert.NotNil(t, config)
}

func TestParseRegisterFlags_Errors(t *testing.T) {
	cmd := newRegisterCommand()
	err := prepareCmdWithFlags(cmd, map[string]string{
		"server-url": "ftp://localhost",
		"username":   "user",
		"password":   "pass",
	})
	require.NoError(t, err)

	_, _, err = parseRegisterFlags(cmd)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported server URL scheme")

	err = prepareCmdWithFlags(cmd, map[string]string{
		"server-url": "http://localhost:8080",
		"username":   "",
		"password":   "pass",
	})
	require.NoError(t, err)

	_, _, err = parseRegisterFlags(cmd)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "username cannot be empty")

	err = prepareCmdWithFlags(cmd, map[string]string{
		"username": "user",
		"password": "",
	})
	require.NoError(t, err)

	_, _, err = parseRegisterFlags(cmd)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "password cannot be empty")
}

func TestRunRegisterApp_NoClientConfigured(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	config := &configs.ClientConfig{}
	creds := &models.Credentials{Username: "u", Password: "p"}

	token, err := runRegisterApp(ctx, config, creds)
	assert.Error(t, err)
	assert.Empty(t, token)
}

func TestParseLoginFlags_Flags(t *testing.T) {
	cmd := newLoginCommand()
	err := prepareCmdWithFlags(cmd, map[string]string{
		"server-url": "http://localhost:8080",
		"username":   "loginflaguser",
		"password":   "loginflagpass",
	})
	require.NoError(t, err)

	config, creds, err := parseLoginFlags(cmd)
	require.NoError(t, err)
	assert.Equal(t, "loginflaguser", creds.Username)
	assert.Equal(t, "loginflagpass", creds.Password)
	assert.NotNil(t, config)
}

func TestParseLoginFlags_Errors(t *testing.T) {
	cmd := newLoginCommand()
	err := prepareCmdWithFlags(cmd, map[string]string{
		"server-url": "xyz://host",
		"username":   "user",
		"password":   "pass",
	})
	require.NoError(t, err)

	_, _, err = parseLoginFlags(cmd)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported server URL scheme")

	err = prepareCmdWithFlags(cmd, map[string]string{
		"server-url": "http://host",
		"username":   "",
		"password":   "pass",
	})
	require.NoError(t, err)

	_, _, err = parseLoginFlags(cmd)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "username cannot be empty")

	err = prepareCmdWithFlags(cmd, map[string]string{
		"username": "user",
		"password": "",
	})
	require.NoError(t, err)

	_, _, err = parseLoginFlags(cmd)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "password cannot be empty")
}

func TestRunLoginApp_NoClientConfigured(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	config := &configs.ClientConfig{}
	creds := &models.Credentials{Username: "u", Password: "p"}

	token, err := runLoginApp(ctx, config, creds)
	assert.Error(t, err)
	assert.Empty(t, token)
}
