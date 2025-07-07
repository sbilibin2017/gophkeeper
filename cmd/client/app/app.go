package app

import "github.com/spf13/cobra"

var (
	use   = "gophkeeper"
	short = "GophKeeper — CLI-инструмент для безопасного управления личными данными"
	long  = `GophKeeper — CLI-инструмент для безопасного управления личными данными.

Использование:
  gophkeeper [команда] [флаги]

Доступные команды:
  build-info       Показать информацию о сборке: платформу, версию, дату и коммит  
  register         Зарегистрировать нового пользователя
  login            Аутентифицировать существующего пользователя  
  add              Добавить новые данные/секреты из файла или интерактивно
  get              Получить конкретные данные/секрет с сервера
  list             Список сохранённых секретов с фильтрацией и сортировкой
  sync             Синхронизировать клиент с сервером и разрешить конфликты  

Используйте "gophkeeper [команда] --help" для получения дополнительной информации о команде.`
)

// NewAppCommand создает корневую команду "gophkeeper" и добавляет все дочерние команды.
func NewAppCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   use,
		Short: short,
		Long:  long,
	}

	cmd.AddCommand(newBuildInfoCommand())
	cmd.AddCommand(newRegisterCommand())
	cmd.AddCommand(newLoginCommand())
	cmd.AddCommand(newAddCommand())
	cmd.AddCommand(newGetCommand())
	cmd.AddCommand(newListCommand())
	cmd.AddCommand(newSyncCommand())

	return cmd
}
