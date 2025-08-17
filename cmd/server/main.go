package main

import (
	"log"

	"github.com/sbilibin2017/gophkeeper/cmd/server/app"
)

// @title           GophKeeper API
// @version         1.0
// @description     API сервер для управления секретами.
// @host            localhost:8080
// @BasePath        /api/v1
// @schemes         http
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Bearer токен для авторизации в формате: "Bearer {token}"
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
