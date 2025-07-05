package app

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	buildPlatform string
	buildVersion  string
	buildDate     string
	buildCommit   string
)

// NewBuildInfoCommand returns a cobra.Command that prints build info.
func newBuildInfoCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "build-info",
		Short: "Show build platform, version, date and commit",
		RunE: func(cmd *cobra.Command, args []string) error {
			printBuildInfo()
			return nil
		},
	}
}

func printBuildInfo() {
	if buildPlatform == "" {
		buildPlatform = "N/A"
	}
	if buildVersion == "" {
		buildVersion = "N/A"
	}
	if buildDate == "" {
		buildDate = "N/A"
	}
	if buildCommit == "" {
		buildCommit = "N/A"
	}
	fmt.Printf("Build platform: %s\n", buildPlatform)
	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build date: %s\n", buildDate)
	fmt.Printf("Build commit: %s\n", buildCommit)
}
