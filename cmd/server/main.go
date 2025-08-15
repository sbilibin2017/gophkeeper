package main

import (
	"log"

	appServer "github.com/sbilibin2017/gophkeeper/internal/apps/server"
	cliServer "github.com/sbilibin2017/gophkeeper/internal/cli/server"
)

func main() {
	err := run()
	if err != nil {
		log.Fatal(err)
	}
}

func run() error {
	cmd := cliServer.NewCommand(appServer.RunHTTP)
	return cmd.Execute()
}
