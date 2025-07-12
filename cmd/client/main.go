package main

import (
	"context"
	"log"

	"github.com/sbilibin2017/gophkeeper/cmd/client/app"
)

func main() {
	err := app.Execute(context.Background())
	if err != nil {
		log.Fatal(err)
	}

}
