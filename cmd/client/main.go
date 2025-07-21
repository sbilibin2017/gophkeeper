package main

import (
	"log"

	"github.com/sbilibin2017/gophkeeper/internal/apps/client"
	"github.com/sbilibin2017/gophkeeper/internal/apps/client/auth/login"
	"github.com/sbilibin2017/gophkeeper/internal/apps/client/auth/logout"
	"github.com/sbilibin2017/gophkeeper/internal/apps/client/auth/register"
	"github.com/sbilibin2017/gophkeeper/internal/apps/client/bankcard"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	rootCmd := client.NewRootCommand()

	register.RegisterCommand(rootCmd, register.RunHTTP, register.RunGRPC)
	login.RegisterCommand(rootCmd, login.RunHTTP, login.RunGRPC)
	logout.RegisterCommand(rootCmd, logout.RunHTTP, logout.RunGRPC)

	bankcard.RegisterAddCommand(rootCmd, bankcard.RunAdd)

	return rootCmd.Execute()
}
