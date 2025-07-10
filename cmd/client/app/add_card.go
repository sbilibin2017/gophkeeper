package app

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/sbilibin2017/gophkeeper/cmd/client/app/config"
	"github.com/sbilibin2017/gophkeeper/cmd/client/app/flags"
	"github.com/spf13/cobra"
)

var (
	cardNumber      string         // номер банковской карты.
	cardExp         string         // срок действия карты (expiry date).
	cardCVV         string         // CVV код карты.
	cardToken       string         // токен авторизации для запроса к серверу.
	cardServerURL   string         // URL сервера для отправки данных.
	cardInteractive bool           // флаг интерактивного режима ввода.
	cardMeta        flags.MetaFlag // метаданные в формате ключ=значение.
)

// newAddCardCommand создает команду для добавления данных банковской карты с метаданными.
func newAddCardCommand() *cobra.Command {
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
  gophkeeper add-card --interactive
`,
		Example: `  gophkeeper add-card --number 4111111111111111 --expiry 12/25 --cvv 123 --meta owner=John --token mytoken --server-url https://example.com
  gophkeeper add-card --interactive`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := parseCardFlags(); err != nil {
				return err
			}

			cfg, err := config.NewConfig(
				config.WithToken(cardToken),
				config.WithServerURL(cardServerURL),
			)
			if err != nil {
				return fmt.Errorf("не удалось создать конфигурацию клиента: %w", err)
			}
			if cfg.ClientConfig.GRPCClient != nil {
				defer cfg.ClientConfig.GRPCClient.Close()
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
	cmd.Flags().StringVar(&cardToken, "token", "", "Токен авторизации")
	cmd.Flags().StringVar(&cardServerURL, "server-url", "", "URL сервера")
	cmd.Flags().BoolVar(&cardInteractive, "interactive", false, "Интерактивный режим ввода")

	return cmd
}

// parseCardFlags обрабатывает флаги и интерактивный ввод для команды add-card.
func parseCardFlags() error {
	if cardInteractive {
		reader := bufio.NewReader(os.Stdin)

		fmt.Print("Введите номер карты: ")
		inputNumber, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		cardNumber = strings.TrimSpace(inputNumber)

		fmt.Print("Введите срок действия (MM/YY): ")
		inputExpiry, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		cardExp = strings.TrimSpace(inputExpiry)

		fmt.Print("Введите CVV код: ")
		inputCVV, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		cardCVV = strings.TrimSpace(inputCVV)

		fmt.Println("Введите метаданные в формате key=value по одному. Пустая строка — завершить:")
		for {
			fmt.Print("> ")
			line, err := reader.ReadString('\n')
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

		fmt.Print("Введите токен авторизации: ")
		inputToken, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		cardToken = strings.TrimSpace(inputToken)

		fmt.Print("Введите URL сервера: ")
		inputServerURL, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		cardServerURL = strings.TrimSpace(inputServerURL)
	}

	if cardNumber == "" || cardExp == "" || cardCVV == "" {
		return fmt.Errorf("параметры number, expiry и cvv обязательны для заполнения")
	}
	if cardToken == "" || cardServerURL == "" {
		return fmt.Errorf("необходимо указать токен и URL сервера")
	}

	return nil
}
