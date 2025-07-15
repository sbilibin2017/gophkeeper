package main

import (
	"log"

	"github.com/sbilibin2017/gophkeeper/cmd/client/app"
)

func main() {
	cmd := app.NewCobraCommand()
	err := cmd.Execute()
	if err != nil {
		log.Fatal(err)
	}
}
