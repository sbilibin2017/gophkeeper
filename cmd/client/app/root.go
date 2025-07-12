package app

import (
	"github.com/sbilibin2017/gophkeeper/cmd/client/app/commands"
	"github.com/spf13/cobra"
)

// NewCommand создает и возвращает корневую команду CLI.
func NewCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "gophkeeper",
		Short: "GophKeeper — безопасный менеджер личных данных",
		Long:  `GophKeeper — это приложение для безопасного хранения и управления личными данными.`,
	}

	// Добавляем подкоманды

	// регистрация + аутентификация
	rootCmd.AddCommand(commands.NewAuthCommand())

	// добавление секретов
	rootCmd.AddCommand(commands.NewAddUsernamePasswordCommand())
	rootCmd.AddCommand(commands.NewAddTextCommand())
	rootCmd.AddCommand(commands.NewAddBinaryCommand())
	rootCmd.AddCommand(commands.NewAddBankCardCommand())

	// получение списков секретов
	rootCmd.AddCommand(commands.NewListUsernamePasswordCommand())
	rootCmd.AddCommand(commands.NewListTextCommand())
	rootCmd.AddCommand(commands.NewListBinaryCommand())
	rootCmd.AddCommand(commands.NewListBankCardCommand())

	return rootCmd
}
