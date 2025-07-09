package app

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestNewAppCommand(t *testing.T) {
	cmd := NewAppCommand()

	// Check main command usage and description
	assert.Equal(t, "gophkeeper", cmd.Use)

	// Build a map of subcommand names for easy lookup
	subcommands := make(map[string]*cobra.Command)
	for _, c := range cmd.Commands() {
		subcommands[c.Name()] = c
	}

	// Check if the "install" subcommand is registered
	installCmd, exists := subcommands["install"]
	assert.True(t, exists, "Expected 'install' command to be registered")

	// Optional: check install command usage or description
	assert.Equal(t, "install", installCmd.Name())

}
