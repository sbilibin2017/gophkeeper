package main

import (
	"log"

	"github.com/sbilibin2017/gophkeeper/cmd/client/app"
)

func main() {
	cmd := app.NewAppCommand()

	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
