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

	register.RegisterRegisterCommand(rootCmd, register.RunRegisterHTTP, register.RunRegisterGRPC)
	login.RegisterLoginCommand(rootCmd, login.RunLoginHTTP, login.RunLoginGRPC)
	logout.RegisterLogoutCommand(rootCmd, logout.RunLogoutHTTP, logout.RunLogoutGRPC)

	bankcard.RegisterAddBankCardCommand(rootCmd, bankcard.RunAddBankCard)

	return rootCmd.Execute()
}
