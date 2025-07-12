package parsemeta

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"
)

func ParseMetaInteractive(r io.Reader) (map[string]string, error) {
	meta := make(map[string]string)
	reader := bufio.NewReader(r)

	fmt.Println("Введите метаданные в формате key=value. Для окончания ввода нажмите Enter на пустой строке.")

	for {
		fmt.Print("meta> ")
		line, err := reader.ReadString('\n')

		// Обработка частичной строки при EOF
		if errors.Is(err, io.EOF) && len(line) > 0 {
			// обрезаем пробелы и парсим строку
			line = strings.TrimSpace(line)
			if line == "" {
				break
			}

			parts := strings.SplitN(line, "=", 2)
			if len(parts) != 2 {
				fmt.Println("Некорректный формат. Введите метаданные в формате key=value.")
				break
			}

			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			if key == "" {
				fmt.Println("Ключ не может быть пустым.")
				break
			}

			meta[key] = value
			break
		}

		if err != nil {
			return nil, errors.New("ошибка при вводе метаданных")
		}

		line = strings.TrimSpace(line)
		if line == "" {
			break
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			fmt.Println("Некорректный формат. Введите метаданные в формате key=value.")
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		if key == "" {
			fmt.Println("Ключ не может быть пустым.")
			continue
		}

		meta[key] = value
	}

	return meta, nil
}
