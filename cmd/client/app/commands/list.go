package commands

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

// NewListCommand создает и возвращает команду для получения списка секретов.
// Поддерживается фильтрация по типу секрета с помощью флага --type.
func NewListCommand() *cobra.Command {
	var secretType string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "Получить список секретов",
		Long:  `Команда для получения списка всех секретов с возможностью фильтрации по типу.`,
		Example: `  # Получить все секреты
  gophkeeper list

  # Получить только секреты типа username-password
  gophkeeper list --type username-password`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runList(context.Background(), secretType)
		},
	}

	cmd.Flags().StringVar(&secretType, "type", "", "Тип секрета для фильтрации (username-password, text, binary, bankcard)")

	return cmd
}

// runList — заглушка функции для получения списка секретов.
// ctx — контекст выполнения,
// secretType — фильтр по типу секрета (пустая строка — без фильтра).
func runList(ctx context.Context, secretType string) error {
	// TODO: Реализовать логику получения и вывода списка секретов с фильтрацией по типу.
	fmt.Printf("Получение списка секретов с фильтром по типу: %q\n", secretType)
	return nil
}
