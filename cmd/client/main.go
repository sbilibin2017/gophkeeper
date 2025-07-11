package main

import (
	"log"

	"github.com/sbilibin2017/gophkeeper/cmd/client/commands"
	"github.com/spf13/cobra"
)

func main() {
	err := run()
	if err != nil {
		log.Fatal(err)
	}
}

func run() error {
	var cmd = &cobra.Command{
		Use:   "gophkeeper",
		Short: "GophKeeper CLI — безопасное хранилище паролей и данных",
	}

	cmd.AddCommand(commands.NewRegisterCommand())

	return cmd.Execute()
}
