package app

import (
	"github.com/sbilibin2017/gophkeeper/internal/configs/buildinfo"
	"github.com/spf13/cobra"
)

// These variables are meant to be set at compile time using -ldflags
var (
	buildPlatform string
	buildVersion  string
	buildDate     string
	buildCommit   string
)

// newBuildInfoCommand returns a cobra.Command
// that outputs build information from compile-time variables.
func newBuildInfoCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "build-info",
		Short: "Show build platform, version, date, and commit",
		RunE: func(cmd *cobra.Command, args []string) error {
			bi := buildinfo.NewBuildInfo(
				buildinfo.WithPlatform(buildPlatform),
				buildinfo.WithVersion(buildVersion),
				buildinfo.WithDate(buildDate),
				buildinfo.WithCommit(buildCommit),
			)

			cmd.Println(bi.String())
			return nil
		},
	}
}
