package main

import (
	"log"

	"github.com/sbilibin2017/gophkeeper/cmd/client/app"
)

func main() {
	err := run()
	if err != nil {
		log.Fatal(err)
	}
}

func run() error {
	cmd := app.NewRootCommand()
	cmd.AddCommand(app.NewRegisterCommand())
	return cmd.Execute()
}
