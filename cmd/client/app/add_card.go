package app

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/sbilibin2017/gophkeeper/cmd/client/app/flags"
	"github.com/sbilibin2017/gophkeeper/cmd/client/app/options"
	"github.com/spf13/cobra"
)

var (
	cardNumber string         // номер карты (глобальная)
	cardExp    string         // срок действия карты (глобальная)
	cardCVV    string         // CVV код карты (глобальная)
	cardMeta   flags.MetaFlag // метаданные key=value (глобальная)
)

// newAddCardCommand создаёт команду "add-card" для добавления банковской карты с метаданными.
// Поддерживает как передачу параметров через флаги, так и интерактивный ввод.
func newAddCardCommand() *cobra.Command {
	var (
		token       string
		serverURL   string
		interactive bool
	)

	cmd := &cobra.Command{
		Use:   "add-card",
		Short: "Добавить данные банковской карты с опциональными метаданными",
		Long: `Команда позволяет добавить номер карты, срок действия и CVV,
а также задать метаданные в формате key=value.

Требуется указать номер карты, срок действия и CVV.
Также необходимы токен авторизации и URL сервера.

Поддерживается интерактивный режим для удобного ввода данных.

Пример использования:

  gophkeeper add-card --number 4111111111111111 --expiry 12/25 --cvv 123 --meta owner=John --token mytoken --server-url https://example.com
  gophkeeper add-card --interactive`,
		Example: `  gophkeeper add-card --number 4111111111111111 --expiry 12/25 --cvv 123 --meta owner=John --token mytoken --server-url https://example.com
  gophkeeper add-card --interactive`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := parseCardFlags(&token, &serverURL, &interactive); err != nil {
				return err
			}

			opts, err := options.NewOptions(
				options.WithToken(token),
				options.WithServerURL(serverURL),
			)
			if err != nil {
				return fmt.Errorf("не удалось создать конфигурацию клиента: %w", err)
			}
			if opts.ClientConfig.GRPCClient != nil {
				defer opts.ClientConfig.GRPCClient.Close()
			}

			fmt.Printf("Добавлена карта:\nНомер: %s\nСрок действия: %s\nCVV: %s\nМетаданные: %+v\n",
				cardNumber, cardExp, cardCVV, cardMeta)

			// TODO: реализовать сохранение карты с метаданными

			return nil
		},
	}

	cmd.Flags().StringVar(&cardNumber, "number", "", "Номер карты")
	cmd.Flags().StringVar(&cardExp, "expiry", "", "Срок действия карты (MM/YY)")
	cmd.Flags().StringVar(&cardCVV, "cvv", "", "CVV код")
	cmd.Flags().Var(&cardMeta, "meta", "Метаданные в формате key=value (можно указывать несколько раз)")

	cmd = options.RegisterTokenFlag(cmd, &token)
	cmd = options.RegisterServerURLFlag(cmd, &serverURL)
	cmd = options.RegisterInteractiveFlag(cmd, &interactive)

	return cmd
}

// parseCardFlags обрабатывает флаги и интерактивный ввод для команды add-card.
// Проверяет обязательные параметры и возвращает ошибку при их отсутствии.
func parseCardFlags(token, serverURL *string, interactive *bool) error {
	if *interactive {
		reader := bufio.NewReader(os.Stdin)
		if err := parseCardFlagsInteractive(reader, token, serverURL); err != nil {
			return err
		}
	}

	if cardNumber == "" || cardExp == "" || cardCVV == "" {
		return fmt.Errorf("параметры number, expiry и cvv обязательны для заполнения")
	}

	if *token == "" || *serverURL == "" {
		return fmt.Errorf("токен и URL сервера должны быть заданы")
	}

	return nil
}

// parseCardFlagsInteractive запрашивает у пользователя необходимые параметры для добавления карты:
// номер карты, срок действия, CVV, метаданные, токен и URL сервера.
func parseCardFlagsInteractive(r *bufio.Reader, token, serverURL *string) error {
	fmt.Print("Введите номер карты: ")
	inputNumber, err := r.ReadString('\n')
	if err != nil {
		return err
	}
	cardNumber = strings.TrimSpace(inputNumber)

	fmt.Print("Введите срок действия (MM/YY): ")
	inputExpiry, err := r.ReadString('\n')
	if err != nil {
		return err
	}
	cardExp = strings.TrimSpace(inputExpiry)

	fmt.Print("Введите CVV код: ")
	inputCVV, err := r.ReadString('\n')
	if err != nil {
		return err
	}
	cardCVV = strings.TrimSpace(inputCVV)

	fmt.Println("Введите метаданные в формате key=value по одному. Пустая строка — завершить:")
	for {
		fmt.Print("> ")
		line, err := r.ReadString('\n')
		if err != nil {
			return err
		}
		line = strings.TrimSpace(line)
		if line == "" {
			break
		}
		if err := cardMeta.Set(line); err != nil {
			return fmt.Errorf("некорректный ввод метаданных: %w", err)
		}
	}

	fmt.Print("Введите токен авторизации (оставьте пустым для использования GOPHKEEPER_TOKEN): ")
	inputToken, err := r.ReadString('\n')
	if err != nil {
		return err
	}
	*token = strings.TrimSpace(inputToken)

	fmt.Print("Введите URL сервера (оставьте пустым для использования GOPHKEEPER_SERVER_URL): ")
	inputServerURL, err := r.ReadString('\n')
	if err != nil {
		return err
	}
	*serverURL = strings.TrimSpace(inputServerURL)

	return nil
}
