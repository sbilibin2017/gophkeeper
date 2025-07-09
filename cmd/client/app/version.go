package app

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	version   = "N/A"
	buildDate = "N/A"
)

func newVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "version",
		Short:   "Show the version and build date",
		Example: "gophkeeper version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("GophKeeper version: %s\nBuild date: %s\n", version, buildDate)
		},
	}
}
