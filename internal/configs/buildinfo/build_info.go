package buildinfo

import "fmt"

// BuildInfo holds information about the software build.
type BuildInfo struct {
	Platform string // Build platform (e.g., linux/amd64)
	Version  string // Build version (e.g., v1.2.3)
	Date     string // Build date (e.g., 2025-07-06)
	Commit   string // Build commit hash
}

// NewBuildInfo creates a new BuildInfo instance applying the given options.
// If an option is not set, default values "N/A" are used.
func NewBuildInfo(opts ...Opt) *BuildInfo {
	b := &BuildInfo{
		Platform: "N/A",
		Version:  "N/A",
		Date:     "N/A",
		Commit:   "N/A",
	}
	for _, opt := range opts {
		opt(b)
	}
	return b
}

// Opt defines a functional option for configuring BuildInfo.
type Opt func(*BuildInfo)

// WithPlatform sets the build platform.
func WithPlatform(platform string) Opt {
	return func(b *BuildInfo) { b.Platform = platform }
}

// WithVersion sets the build version.
func WithVersion(version string) Opt {
	return func(b *BuildInfo) { b.Version = version }
}

// WithDate sets the build date.
func WithDate(date string) Opt {
	return func(b *BuildInfo) { b.Date = date }
}

// WithCommit sets the build commit hash.
func WithCommit(commit string) Opt {
	return func(b *BuildInfo) { b.Commit = commit }
}

// String returns a formatted string representation of the BuildInfo.
// It implements the fmt.Stringer interface.
func (b *BuildInfo) String() string {
	return fmt.Sprintf(
		"Build platform: %s\nBuild version: %s\nBuild date: %s\nBuild commit: %s",
		b.Platform, b.Version, b.Date, b.Commit,
	)
}
