package app

import (
	"github.com/sbilibin2017/gophkeeper/cmd/client/app/commands"
	"github.com/spf13/cobra"
)

// NewCobraCommand создает и возвращает корневую команду Cobra,
// регистрирует все дочерние команды.
func NewCobraCommand() *cobra.Command {
	root := &cobra.Command{
		Use:   "gophkeeper-client",
		Short: "Клиент для работы с GophKeeper",
	}

	commands.RegisterVersionCommand(root)

	commands.RegisterRegisterCommand(root)
	commands.RegisterLoginCommand(root)

	commands.RegisterAddSecretBankCardCommand(root)
	commands.RegisterAddSecretBinaryCommand(root)
	commands.RegisterAddSecretTextCommand(root)
	commands.RegisterAddSecretUsernamePasswordCommand(root)

	commands.RegisterListSecretCommand(root)

	commands.RegisterSyncCommand(root)

	return root
}
