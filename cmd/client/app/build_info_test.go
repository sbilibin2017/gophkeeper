package app

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewBuildInfoCommand_RunE(t *testing.T) {
	// Override the build variables for the test
	buildPlatform = "test-platform"
	buildVersion = "v0.1.2-test"
	buildDate = "2025-07-06"
	buildCommit = "deadbeef"

	cmd := newBuildInfoCommand()

	// Disable cobra's default error and usage printing to keep test output clean
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true

	// Capture output by redirecting cmd.OutOrStdout
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	// Run the command's RunE function directly
	err := cmd.RunE(cmd, []string{})
	assert.NoError(t, err)

	out := buf.String()

	assert.Contains(t, out, "Build platform: test-platform")
	assert.Contains(t, out, "Build version: v0.1.2-test")
	assert.Contains(t, out, "Build date: 2025-07-06")
	assert.Contains(t, out, "Build commit: deadbeef")
}
