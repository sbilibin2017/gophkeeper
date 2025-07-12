package commands

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

// NewSyncCommand создает и возвращает команду для синхронизации клиента с сервером.
// Флаг --resolve определяет способ разрешения конфликтов ("server" или "client").
// Флаг --interactive включает интерактивный режим.
func NewSyncCommand() *cobra.Command {
	var resolve string
	var interactive bool

	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Синхронизировать клиента с сервером",
		Long:  `Команда для синхронизации данных между клиентом и сервером с опцией разрешения конфликтов.`,
		Example: `  # Синхронизировать с разрешением конфликтов в пользу сервера
  gophkeeper sync --resolve server

  # Синхронизировать с разрешением конфликтов в пользу клиента
  gophkeeper sync --resolve client

  # Синхронизировать в интерактивном режиме
  gophkeeper sync --interactive`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return syncClientServer(context.Background(), resolve, interactive, bufio.NewReader(os.Stdin))
		},
	}

	cmd.Flags().StringVar(&resolve, "resolve", "client", "Способ разрешения конфликтов: server или client")
	cmd.Flags().BoolVar(&interactive, "interactive", false, "Интерактивный режим")

	return cmd
}

// syncClientServer — заглушка логики синхронизации.
// ctx — контекст выполнения,
// resolve — способ разрешения конфликтов,
// interactive — флаг интерактивного режима,
// reader — источник интерактивного ввода.
func syncClientServer(ctx context.Context, resolve string, interactive bool, reader *bufio.Reader) error {
	resolve = strings.ToLower(resolve)
	if resolve != "server" && resolve != "client" {
		return errors.New("флаг --resolve должен быть 'server' или 'client'")
	}

	if interactive {
		fmt.Println("Интерактивный режим синхронизации включен")
		// Здесь можно реализовать интерактивный ввод/подтверждение
	}

	fmt.Printf("Синхронизация с разрешением конфликтов в пользу: %s\n", resolve)
	// Здесь будет основная логика синхронизации

	return nil
}
