package main

import (
	"log"

	"github.com/sbilibin2017/gophkeeper/inernal/apps/server"
)

func main() {
	err := run()
	if err != nil {
		log.Fatal(err)
	}
}

func run() error {
	rootCmd := server.NewCommand()
	return rootCmd.Execute()
}
