package buildinfo

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewBuildInfo_Defaults(t *testing.T) {
	b := NewBuildInfo()

	assert.Equal(t, "N/A", b.Platform)
	assert.Equal(t, "N/A", b.Version)
	assert.Equal(t, "N/A", b.Date)
	assert.Equal(t, "N/A", b.Commit)
}

func TestNewBuildInfo_WithOptions(t *testing.T) {
	b := NewBuildInfo(
		WithPlatform("linux/amd64"),
		WithVersion("v1.2.3"),
		WithDate("2025-07-06"),
		WithCommit("abc123"),
	)

	assert.Equal(t, "linux/amd64", b.Platform)
	assert.Equal(t, "v1.2.3", b.Version)
	assert.Equal(t, "2025-07-06", b.Date)
	assert.Equal(t, "abc123", b.Commit)
}

func TestBuildInfo_String(t *testing.T) {
	b := NewBuildInfo(
		WithPlatform("linux/amd64"),
		WithVersion("v1.2.3"),
		WithDate("2025-07-06"),
		WithCommit("abc123"),
	)

	output := b.String()

	assert.True(t, strings.Contains(output, "Build platform: linux/amd64"))
	assert.True(t, strings.Contains(output, "Build version: v1.2.3"))
	assert.True(t, strings.Contains(output, "Build date: 2025-07-06"))
	assert.True(t, strings.Contains(output, "Build commit: abc123"))
}
