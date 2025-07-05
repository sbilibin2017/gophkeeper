package app

import (
	"bytes"
	"os"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

func TestNewBuildInfoCommand(t *testing.T) {
	// Set build info for test
	buildPlatform = "linux/amd64"
	buildVersion = "v1.2.3"
	buildDate = "2025-07-04"
	buildCommit = "abcdef123456"

	// Capture os.Stdout
	var buf bytes.Buffer
	stdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Run command
	rootCmd := &cobra.Command{Use: "root"}
	rootCmd.AddCommand(newBuildInfoCommand())
	rootCmd.SetArgs([]string{"build-info"})

	err := rootCmd.Execute()
	require.NoError(t, err)

	// Restore and read output
	w.Close()
	os.Stdout = stdout
	_, _ = buf.ReadFrom(r)

	output := buf.String()

	// Assertions
	require.Contains(t, output, "Build platform: linux/amd64")
	require.Contains(t, output, "Build version: v1.2.3")
	require.Contains(t, output, "Build date: 2025-07-04")
	require.Contains(t, output, "Build commit: abcdef123456")
}

func TestBuildInfoCommand_DefaultsToNA(t *testing.T) {
	// Set all values to empty
	buildPlatform = ""
	buildVersion = ""
	buildDate = ""
	buildCommit = ""

	// Capture os.Stdout
	var buf bytes.Buffer
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Run command
	rootCmd := &cobra.Command{Use: "root"}
	rootCmd.AddCommand(newBuildInfoCommand())
	rootCmd.SetArgs([]string{"build-info"})

	err := rootCmd.Execute()
	require.NoError(t, err)

	// Restore and read output
	w.Close()
	os.Stdout = old
	_, _ = buf.ReadFrom(r)

	output := buf.String()

	// Assertions for default "N/A" values
	require.Contains(t, output, "Build platform: N/A")
	require.Contains(t, output, "Build version: N/A")
	require.Contains(t, output, "Build date: N/A")
	require.Contains(t, output, "Build commit: N/A")
}
