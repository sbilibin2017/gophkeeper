// Пакет main реализует CLI-клиент для GophKeeper.
// Поддерживаются команды "register" и "login" с возможностью указания флагов
// для настройки сервера и интерактивного ввода логина/пароля.
package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/sbilibin2017/gophkeeper/cmd/client/commands"
)

// init инициализирует глобальные флаги, доступные для всех CLI-команд.
func init() {
	flag.String("server-url", "http://localhost:8080", "URL сервера GophKeeper")
	flag.Bool("interactive", false, "Запросить ввод логина и пароля в интерактивном режиме")
}

// main — точка входа в приложение.
// Вызывает функцию run() и завершает выполнение в случае ошибки.
func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

// run обрабатывает аргументы командной строки и вызывает соответствующую команду.
// Возвращает ошибку, если команда не указана или не распознана.
func run() error {
	if len(os.Args) < 2 {
		return fmt.Errorf("не указана команда")
	}

	flag.Parse()

	ctx := context.Background()
	cmd := os.Args[1]
	args := os.Args[2:]
	flags := parseFlags(flag.CommandLine)
	envs := os.Environ()
	reader := bufio.NewReader(os.Stdin)

	switch cmd {
	case "register":
		return commands.RegisterCommand(ctx, args, flags, envs, reader)
	case "login":
		return commands.LoginCommand(ctx, args, flags, envs, reader)
	default:
		return fmt.Errorf("неизвестная команда: %s", cmd)
	}
}

// parseFlags извлекает переданные пользователем флаги из FlagSet.
// Возвращает карту с именами флагов и их строковыми значениями.
func parseFlags(fs *flag.FlagSet) map[string]string {
	flags := make(map[string]string)
	fs.Visit(func(f *flag.Flag) {
		flags[f.Name] = f.Value.String()
	})
	return flags
}
