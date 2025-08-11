package main

import (
	"log"

	"github.com/sbilibin2017/gophkeeper/internal/client"
	"github.com/sbilibin2017/gophkeeper/internal/client/add/bankcard"
	"github.com/sbilibin2017/gophkeeper/internal/client/add/binary"
	"github.com/sbilibin2017/gophkeeper/internal/client/add/text"
	"github.com/sbilibin2017/gophkeeper/internal/client/add/user"
	"github.com/sbilibin2017/gophkeeper/internal/client/auth/login"
	"github.com/sbilibin2017/gophkeeper/internal/client/auth/register"
	"github.com/sbilibin2017/gophkeeper/internal/client/list"
	"github.com/sbilibin2017/gophkeeper/internal/client/sync"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	cmd := client.NewCommand()

	cmd.AddCommand(register.NewCommand())
	cmd.AddCommand(login.NewCommand())

	cmd.AddCommand(bankcard.NewCommand())
	cmd.AddCommand(binary.NewCommand())
	cmd.AddCommand(text.NewCommand())
	cmd.AddCommand(user.NewCommand())
	cmd.AddCommand(list.NewCommand())
	cmd.AddCommand(sync.NewCommand())

	return cmd.Execute()
}
