package main

import (
	"log"

	"github.com/sbilibin2017/gophkeeper/internal/apps"
	"github.com/sbilibin2017/gophkeeper/internal/commands"
	"github.com/sbilibin2017/gophkeeper/internal/configs/device"
)

func main() {
	err := run()
	if err != nil {
		log.Fatal(err)
	}
}

func run() error {
	cmd := commands.NewRootCommand()
	cmd.AddCommand(
		commands.NewRegisterCommand(
			apps.RunClientRegisterHTTP,
			device.GetDeviceID,
		),
	)
	return cmd.Execute()
}
