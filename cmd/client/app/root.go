package app

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/sbilibin2017/gophkeeper/cmd/client/app/commands"
)

// init инициализирует глобальные флаги, доступные для всех CLI-команд.
func init() {
	flag.String("server-url", "http://localhost:8080", "URL сервера GophKeeper")
	flag.Bool("interactive", false, "Запросить ввод логина и пароля в интерактивном режиме")
	flag.String("meta", "", "Произвольная метаинформация в формате key1=value1,key2=value2")
}

// Execute обрабатывает команду, переданную через аргументы командной строки.
func Execute(ctx context.Context) error {
	// Проверяем, что в аргументах командной строки есть хотя бы команда
	if len(os.Args) < 2 {
		return fmt.Errorf("не указана команда")
	}

	// Парсим глобальные флаги из os.Args (начиная со второго аргумента)
	flag.Parse()

	// Первая позиционная аргумента - команда (например, register, login)
	cmd := os.Args[1]

	// Остальные позиционные аргументы — параметры команды
	args := os.Args[2:]

	// Извлекаем флаги в map[string]string для удобной работы
	flags := parseFlags(flag.CommandLine)

	// Обрабатываем флаг meta: парсим строку в структуру, проверяем формат и сериализуем обратно
	if err := extractMeta(flags); err != nil {
		return fmt.Errorf("ошибка парсинга meta: %w", err)
	}

	// Получаем переменные окружения для передачи в команды
	envs := os.Environ()

	// Создаём reader для чтения пользовательского ввода (если нужен интерактивный режим)
	reader := bufio.NewReader(os.Stdin)

	// В зависимости от команды вызываем соответствующую функцию из commands
	switch cmd {
	case "register":
		return commands.Register(ctx, args, *flags, envs, reader)
	case "login":
		return commands.Login(ctx, args, *flags, envs, reader)
	default:
		return fmt.Errorf("неизвестная команда: %s", cmd)
	}
}

// parseFlags извлекает значения флагов из переданного FlagSet
// и возвращает их в виде словаря (имя флага -> значение).
func parseFlags(fs *flag.FlagSet) *map[string]string {
	flags := make(map[string]string)
	fs.Visit(func(f *flag.Flag) {
		flags[f.Name] = f.Value.String()
	})
	return &flags
}

// extractMeta извлекает из flags строку метаинформации по ключу "meta",
// проверяет её формат, парсит в map и записывает обратно в flags["meta"] в виде JSON-строки.
func extractMeta(flags *map[string]string) error {
	metaStr, ok := (*flags)["meta"]
	if !ok || metaStr == "" {
		return nil
	}

	metaMap := make(map[string]string)

	pairs := strings.Split(metaStr, ",")
	for _, pair := range pairs {
		if !strings.Contains(pair, "=") {
			return fmt.Errorf("неверный формат метаинформации: %s", pair)
		}
		kv := strings.SplitN(pair, "=", 2)
		metaMap[kv[0]] = kv[1]
	}

	jsonBytes, err := json.Marshal(metaMap)
	if err != nil {
		return fmt.Errorf("ошибка сериализации meta: %w", err)
	}

	(*flags)["meta"] = string(jsonBytes)

	return nil
}
