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

func init() {
	flag.String("server-url", "http://localhost:8080", "URL сервера GophKeeper")
	flag.Bool("interactive", false, "Запросить ввод логина и пароля в интерактивном режиме")
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

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
		return commands.RegisterCommand(
			ctx, args, flags, envs, reader,
		)
	default:
		return fmt.Errorf("неизвестная команда: %s", cmd)
	}
}

func parseFlags(fs *flag.FlagSet) map[string]string {
	flags := make(map[string]string)
	fs.Visit(func(f *flag.Flag) {
		flags[f.Name] = f.Value.String()
	})
	return flags
}
