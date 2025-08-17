package main

import (
	"log"

	"github.com/sbilibin2017/gophkeeper/internal/apps"
	"github.com/sbilibin2017/gophkeeper/internal/commands"
)

// @title           GophKeeper API
// @version         1.0
// @description     API сервер для управления секретами.
// @host            localhost:8080
// @BasePath        /
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
	cmd := commands.NewServerCommand(apps.RunServerHTTP)
	return cmd.Execute()
}
