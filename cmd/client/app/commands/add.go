package commands

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/sbilibin2017/gophkeeper/cmd/client/app/commands/db"
	"github.com/sbilibin2017/gophkeeper/cmd/client/app/commands/parsemeta"
	"github.com/sbilibin2017/gophkeeper/internal/configs"
	"github.com/sbilibin2017/gophkeeper/internal/models"

	"github.com/sbilibin2017/gophkeeper/internal/services"
	"github.com/spf13/cobra"
)

// NewAddUsernamePasswordCommand создает и возвращает команду для добавления username и password в секреты.
func NewAddUsernamePasswordCommand() *cobra.Command {
	var interactive bool
	var metaFlags []string

	cmd := &cobra.Command{
		Use:   "add-username-password [username password]",
		Short: "Добавить username и password в секреты",
		Long:  `Добавление username и password в безопасное хранилище. Аргументы или интерактивный ввод.`,
		Args:  cobra.MaximumNArgs(2),
		Example: `  # Добавить username и password через аргументы
  gophkeeper add-secret-username-password myusername mypassword --meta env=prod --meta version=1

  # Добавить username и password интерактивно
  gophkeeper add-secret-username-password --interactive --meta env=prod`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return addUsernamePassword(
				context.Background(),
				args,
				interactive,
				metaFlags,
				bufio.NewReader(os.Stdin),
			)
		},
	}

	cmd.Flags().BoolVar(&interactive, "interactive", false, "Интерактивный ввод username и password")
	cmd.Flags().StringArrayVar(&metaFlags, "meta", nil, "Дополнительные метаданные в формате key=value (можно указывать несколько)")

	return cmd
}

func addUsernamePassword(ctx context.Context, args []string, interactive bool, metaFlags []string, reader *bufio.Reader) error {
	config, err := configs.NewClientConfig(
		configs.WithDB("gophkeeper.db"),
	)
	if err != nil {
		return errors.New("клиент не был сконфигурирован")
	}

	err = db.Migrate(config.DB)
	if err != nil {
		return fmt.Errorf("не удалось выполнить миграцию базы данных: %w", err)
	}

	meta := parsemeta.ParseMeta(metaFlags)
	if interactive {
		metaInteractive, err := parsemeta.ParseMetaInteractive(reader)
		if err != nil {
			return err
		}
		for k, v := range metaInteractive {
			meta[k] = v
		}
	}

	var username, password string
	if interactive {
		fmt.Print("Введите username: ")
		usernameRaw, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		username = strings.TrimSpace(usernameRaw)

		fmt.Print("Введите password: ")
		passwordRaw, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		password = strings.TrimSpace(passwordRaw)
	} else {
		if len(args) < 2 {
			return errors.New("требуется указать username и password либо использовать --interactive")
		}
		username = args[0]
		password = args[1]
	}

	secret := &models.UsernamePassword{
		Username: username,
		Password: password,
		Meta:     meta,
	}

	err = services.AddLoginPassword(ctx, config.DB, secret)
	if err != nil {
		return fmt.Errorf("не удалось добавить username и password: %w", err)
	}

	return nil
}

// NewAddTextCommand создает и возвращает команду для добавления произвольных текстовых данных в секреты.
func NewAddTextCommand() *cobra.Command {
	var interactive bool
	var metaFlags []string

	cmd := &cobra.Command{
		Use:   "add-text [content]",
		Short: "Добавить текстовые данные в секреты",
		Long:  `Добавление произвольных текстовых данных в безопасное хранилище. Аргумент или интерактивный ввод.`,
		Args:  cobra.MaximumNArgs(1),
		Example: `  # Добавить текст через аргумент
  gophkeeper add-secret-text "some text content" --meta source=manual

  # Добавить текст интерактивно
  gophkeeper add-secret-text --interactive --meta source=manual`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return addText(
				context.Background(),
				args,
				interactive,
				metaFlags,
				bufio.NewReader(os.Stdin),
			)
		},
	}

	cmd.Flags().BoolVar(&interactive, "interactive", false, "Интерактивный ввод текста")
	cmd.Flags().StringArrayVar(&metaFlags, "meta", nil, "Дополнительные метаданные в формате key=value (можно указывать несколько)")

	return cmd
}

func addText(ctx context.Context, args []string, interactive bool, metaFlags []string, reader *bufio.Reader) error {
	config, err := configs.NewClientConfig(
		configs.WithDB("gophkeeper.db"),
	)
	if err != nil {
		return errors.New("клиент не был сконфигурирован")
	}

	err = db.Migrate(config.DB)
	if err != nil {
		return errors.New("клиент не был сконфигурирован")
	}

	meta := parsemeta.ParseMeta(metaFlags)
	if interactive {
		metaInteractive, err := parsemeta.ParseMetaInteractive(reader)
		if err != nil {
			return err
		}
		for k, v := range metaInteractive {
			meta[k] = v
		}
	}

	var content string
	if interactive {
		fmt.Println("Введите текст (окончание ввода — пустая строка):")
		lines := []string{}
		for {
			lineRaw, err := reader.ReadString('\n')
			if err != nil {
				return err
			}
			line := strings.TrimSpace(lineRaw)
			if line == "" {
				break
			}
			lines = append(lines, line)
		}
		content = strings.Join(lines, "\n")
	} else {
		if len(args) < 1 {
			return errors.New("требуется указать текст либо использовать --interactive")
		}
		content = args[0]
	}

	text := &models.Text{
		Content: content,
		Meta:    meta,
	}

	err = services.AddText(ctx, config.DB, text)
	if err != nil {
		return fmt.Errorf("не удалось добавить текст: %w", err)
	}

	return nil
}

// NewAddBinaryCommand создает и возвращает команду для добавления бинарных данных в секреты.
func NewAddBinaryCommand() *cobra.Command {
	var metaFlags []string

	cmd := &cobra.Command{
		Use:   "add-binary [filepath]",
		Short: "Добавить бинарные данные в секреты",
		Long:  `Добавление произвольных бинарных данных из файла в безопасное хранилище.`,
		Args:  cobra.ExactArgs(1),
		Example: `  # Добавить бинарные данные из файла
  gophkeeper add-secret-binary /path/to/file --meta type=image`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return addBinary(
				context.Background(),
				args,
				metaFlags,
			)
		},
	}

	cmd.Flags().StringArrayVar(&metaFlags, "meta", nil, "Дополнительные метаданные в формате key=value (можно указывать несколько)")

	return cmd
}

func addBinary(ctx context.Context, args []string, metaFlags []string) error {
	config, err := configs.NewClientConfig(
		configs.WithDB("gophkeeper.db"),
	)
	if err != nil {
		return errors.New("клиент не был сконфигурирован")
	}

	err = db.Migrate(config.DB)
	if err != nil {
		return errors.New("клиент не был сконфигурирован")
	}

	meta := parsemeta.ParseMeta(metaFlags)

	filepath := args[0]
	data, err := os.ReadFile(filepath)
	if err != nil {
		return fmt.Errorf("не удалось прочитать файл: %w", err)
	}

	bin := &models.Binary{
		Data: data,
		Meta: meta,
	}

	err = services.AddBinary(ctx, config.DB, bin)
	if err != nil {
		return fmt.Errorf("не удалось добавить бинарные данные: %w", err)
	}

	return nil
}

// NewAddBankCardCommand создает и возвращает команду для добавления данных банковской карты в секреты.
func NewAddBankCardCommand() *cobra.Command {
	var interactive bool
	var metaFlags []string

	cmd := &cobra.Command{
		Use:   "add-bank-card",
		Short: "Добавить данные банковской карты в секреты",
		Long:  `Добавление данных банковской карты в безопасное хранилище с возможностью интерактивного ввода.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return addBankCard(
				context.Background(),
				interactive,
				metaFlags,
				bufio.NewReader(os.Stdin),
			)
		},
	}

	cmd.Flags().BoolVar(&interactive, "interactive", false, "Интерактивный ввод данных банковской карты")
	cmd.Flags().StringArrayVar(&metaFlags, "meta", nil, "Дополнительные метаданные в формате key=value (можно указывать несколько)")

	return cmd
}

func addBankCard(ctx context.Context, interactive bool, metaFlags []string, reader *bufio.Reader) error {
	config, err := configs.NewClientConfig(
		configs.WithDB("gophkeeper.db"),
	)
	if err != nil {
		return errors.New("клиент не был сконфигурирован")
	}

	err = db.Migrate(config.DB)
	if err != nil {
		return errors.New("клиент не был сконфигурирован")
	}

	meta := parsemeta.ParseMeta(metaFlags)
	if interactive {
		metaInteractive, err := parsemeta.ParseMetaInteractive(reader)
		if err != nil {
			return err
		}
		for k, v := range metaInteractive {
			meta[k] = v
		}
	}

	var number, expiry, cvv string
	if interactive {
		fmt.Print("Введите номер карты: ")
		numberRaw, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		number = strings.TrimSpace(numberRaw)

		fmt.Print("Введите срок действия карты (MM/YY): ")
		expiryRaw, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		expiry = strings.TrimSpace(expiryRaw)

		fmt.Print("Введите CVV: ")
		cvvRaw, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		cvv = strings.TrimSpace(cvvRaw)
	} else {
		return errors.New("требуется использовать --interactive для ввода данных банковской карты")
	}

	card := &models.BankCard{
		Number: number,
		Expiry: expiry,
		CVV:    cvv,
		Meta:   meta,
	}

	err = services.AddBankCard(ctx, config.DB, card)
	if err != nil {
		return fmt.Errorf("не удалось добавить данные банковской карты: %w", err)
	}

	return nil
}
