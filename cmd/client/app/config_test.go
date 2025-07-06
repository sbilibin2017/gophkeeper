package app

import (
	"bytes"
	"testing"

	"github.com/sbilibin2017/gophkeeper/internal/configs/file"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClientConfigSetGetUnsetCommands(t *testing.T) {
	const key, value = "test_key", "test_value"

	// Cleanup key after test run
	defer func() {
		_ = file.UnsetConfigValue(key)
	}()

	t.Run("SetConfig", func(t *testing.T) {
		buf := new(bytes.Buffer)
		cmd := newClientConfigSetCommand()
		cmd.SetOut(buf)
		cmd.SetArgs([]string{key, value})

		err := cmd.Execute()
		require.NoError(t, err)
		assert.Contains(t, buf.String(), "Set")
	})

	t.Run("GetConfigKey", func(t *testing.T) {
		// Ensure key is set first
		err := file.SetConfigValue(key, value)
		require.NoError(t, err)

		buf := new(bytes.Buffer)
		cmd := newClientConfigGetCommand()
		cmd.SetOut(buf)
		cmd.SetArgs([]string{key})

		err = cmd.Execute()
		require.NoError(t, err)
		output := buf.String()
		assert.Contains(t, output, key)
		assert.Contains(t, output, value)
	})

	t.Run("GetConfigAll", func(t *testing.T) {
		// Ensure key is set first
		err := file.SetConfigValue(key, value)
		require.NoError(t, err)

		buf := new(bytes.Buffer)
		cmd := newClientConfigGetCommand()
		cmd.SetOut(buf)
		cmd.SetArgs([]string{}) // no args to get all config

		err = cmd.Execute()
		require.NoError(t, err)
		output := buf.String()
		assert.Contains(t, output, key)
		assert.Contains(t, output, value)
	})

	t.Run("UnsetConfig", func(t *testing.T) {
		// Ensure key is set first
		err := file.SetConfigValue(key, value)
		require.NoError(t, err)

		buf := new(bytes.Buffer)
		cmd := newClientConfigUnsetCommand()
		cmd.SetOut(buf)
		cmd.SetArgs([]string{key})

		err = cmd.Execute()
		require.NoError(t, err)
		assert.Contains(t, buf.String(), "Unset")

		// Confirm key is gone
		val, ok, err := file.GetConfigValue(key)
		require.NoError(t, err)
		assert.False(t, ok)
		assert.Empty(t, val)
	})
}

func TestClientConfigGetCommand_KeyNotFound(t *testing.T) {
	const missingKey = "non_existent_key"
	buf := new(bytes.Buffer)
	cmd := newClientConfigGetCommand()
	cmd.SetOut(buf)
	cmd.SetArgs([]string{missingKey})

	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}
