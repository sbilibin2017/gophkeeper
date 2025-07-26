package cli

import (
	"log"

	"github.com/spf13/cobra"
)

func Run(cmd *cobra.Command) int {
	err := cmd.Execute()
	if err != nil {
		log.Println(err)
		return 1
	}
	return 0
}
