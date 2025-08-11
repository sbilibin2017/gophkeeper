package client

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewCommand(t *testing.T) {
	cmd := NewCommand()

	assert.NotNil(t, cmd, "command should not be nil")
	assert.Equal(t, "gophkeeper", cmd.Use, "command Use should be 'gophkeeper'")
	assert.Contains(t, cmd.Short, "Gophkeeper", "command Short description should contain 'Gophkeeper'")
	assert.Contains(t, cmd.Long, "securely storing", "command Long description should mention 'securely storing'")
}
