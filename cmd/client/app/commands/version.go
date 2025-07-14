package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	// Version содержит строку с версией клиента.
	// Устанавливается при сборке через флаг компилятора -ldflags.
	Version = "N/A"

	// BuildDate содержит дату и время сборки клиента.
	// Устанавливается при сборке через флаг компилятора -ldflags.
	BuildDate = "N/A"
)

// RegisterVersionCommand регистрирует команду CLI "version" для отображения версии клиента.
//
// Пример использования:
//
//	gophkeeper-client version
func RegisterVersionCommand(root *cobra.Command) {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Показать версию и дату сборки клиента",
		Long:  `Отображает информацию о версии клиента GophKeeper и дате его сборки.`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("GophKeeper client")
			fmt.Printf("Version: %s\n", Version)
			fmt.Printf("Build Date: %s\n", BuildDate)
		},
	}

	root.AddCommand(cmd)
}
