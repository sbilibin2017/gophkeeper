package main

import (
	"log"

	"github.com/sbilibin2017/gophkeeper/internal/server"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	cmd := server.NewCommand()
	return cmd.Execute()
}
