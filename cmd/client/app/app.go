package app

import (
	"github.com/spf13/cobra"
)

// NewAppCommand создаёт корневую команду CLI-приложения GophKeeper.
// Включает все доступные подкоманды для управления приватными данными.
func NewAppCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gophkeeper",
		Short: "GophKeeper — CLI менеджер для приватных данных",
		Long:  "GophKeeper — CLI инструмент для безопасного управления вашими приватными данными: логинами, текстами, файлами, картами и другими.",
	}

	// Добавление подкоманд
	cmd.AddCommand(newVersionCommand())          // Вывод версии и информации о сборке
	cmd.AddCommand(newConfigCommand())           // Настройка параметров клиента (токен, URL сервера)
	cmd.AddCommand(newRegisterCommand())         // Регистрация нового пользователя
	cmd.AddCommand(newLoginCommand())            // Вход существующего пользователя
	cmd.AddCommand(newAddLoginPasswordCommand()) // Добавление пары логин/пароль
	cmd.AddCommand(newAddTextCommand())          // Добавление произвольного текста
	cmd.AddCommand(newAddBinaryCommand())        // Добавление бинарных данных из файла
	cmd.AddCommand(newAddCardCommand())          // Добавление данных банковской карты
	cmd.AddCommand(newListCommand())             // Вывод списка сохранённых секретов
	cmd.AddCommand(newSyncCommand())             // Синхронизация локальных данных с сервером

	return cmd
}
