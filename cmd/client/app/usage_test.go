package app

import (
	"bytes"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

func TestNewUsageCommand(t *testing.T) {
	rootCmd := &cobra.Command{Use: "root"}
	rootCmd.AddCommand(newUsageCommand())

	// Set up a buffer to capture output
	output := new(bytes.Buffer)
	rootCmd.SetOut(output)
	rootCmd.SetArgs([]string{"usage"})

	// Execute the command
	err := rootCmd.Execute()

	require.NoError(t, err)
	require.Contains(t, output.String(), "Show usage information")
}
