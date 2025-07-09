// Package app содержит реализацию CLI-приложения GophKeeper.
//
// GophKeeper — это CLI-инструмент для безопасного управления личными данными.
// Основная команда предоставляет доступ к подкомандам.
package app

import "github.com/spf13/cobra"

var (
	use   = "gophkeeper"
	short = "GophKeeper — CLI-инструмент для безопасного управления личными данными"
	long  = `GophKeeper — CLI-инструмент для безопасного управления личными данными.
Использование:
  gophkeeper [команда] [флаги]
Доступные команды:
  install       Установить клиент для текущей платформы
Используйте "gophkeeper [команда] --help" для получения дополнительной информации о команде.`
)

// NewAppCommand создаёт и возвращает основную команду CLI-приложения GophKeeper.
//
// Команда включает подкоманды.
func NewAppCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   use,
		Short: short,
		Long:  long,
	}

	cmd.AddCommand(newInstallCommand())

	return cmd
}
