package main

import (
	"log"

	"github.com/sbilibin2017/gophkeeper/cmd/server/app"
)

func main() {
	err := run()
	if err != nil {
		log.Fatal(err)
	}
}

func run() error {
	cmd := app.NewCommand()
	return cmd.Execute()
}
