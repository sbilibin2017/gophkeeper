package app

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	buildVersion = "N/A" // версия сервиса
	buildDate    = "N/A" // дата сборки сервиса
)

// newVersionCommand создаёт и возвращает команду "version" для CLI.
func newVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Выводит текущую версию приложения GophKeeper и дату сборки",
		Long: `Команда version выводит текущую версию приложения GophKeeper и дату сборки.

Значения версии и даты сборки задаются с помощью флагов компиляции (-ldflags) при сборке приложения.

Пример сборки с указанием версии и даты:
  go build -ldflags "-X 'github.com/sbilibin2017/gophkeeper/cmd/client/app.buildVersion=1.0.0' -X 'github.com/sbilibin2017/gophkeeper/cmd/client/app.buildDate=2025-07-10'" -o gophkeeper
`,
		Example: "gophkeeper version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Fprintf(cmd.OutOrStdout(), "version: %s\ndate: %s\n", buildVersion, buildDate)
		},
	}
}
