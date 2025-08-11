package main

import (
	"log"

	"github.com/sbilibin2017/gophkeeper/internal/client"
	"github.com/sbilibin2017/gophkeeper/internal/client/login"
	"github.com/sbilibin2017/gophkeeper/internal/client/register"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	cmd := client.NewCommand()
	cmd.AddCommand(register.NewCommand())
	cmd.AddCommand(login.NewCommand())
	return cmd.Execute()
}
