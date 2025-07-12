package app

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/sbilibin2017/gophkeeper/cmd/client/app/commands"
)

// init инициализирует глобальные флаги, доступные для всех CLI-команд.
func init() {
	flag.String("server-url", "http://localhost:8080", "URL сервера GophKeeper")
	flag.Bool("interactive", false, "Запросить ввод логина и пароля в интерактивном режиме")
}

// Execute обрабатывает команду, переданную через аргументы командной строки.
// Проверяет наличие команды, парсит флаги и вызывает соответствующую функцию команды.
// Поддерживаются команды: register, login.
// Возвращает ошибку, если команда не указана или неизвестна.
func Execute(ctx context.Context) error {
	if len(os.Args) < 2 {
		return fmt.Errorf("не указана команда")
	}

	flag.Parse()

	cmd := os.Args[1]
	args := os.Args[2:]
	flags := parseFlags(flag.CommandLine)
	envs := os.Environ()
	reader := bufio.NewReader(os.Stdin)

	switch cmd {
	case "register":
		return commands.Register(ctx, args, flags, envs, reader)
	case "login":
		return commands.Login(ctx, args, flags, envs, reader)
	default:
		return fmt.Errorf("неизвестная команда: %s", cmd)
	}
}

// parseFlags извлекает значения флагов из переданного FlagSet
// и возвращает их в виде словаря (имя флага -> значение).
func parseFlags(fs *flag.FlagSet) map[string]string {
	flags := make(map[string]string)
	fs.Visit(func(f *flag.Flag) {
		flags[f.Name] = f.Value.String()
	})
	return flags
}
