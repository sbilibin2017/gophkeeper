package parsemeta

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"
)

func ParseMeta(meta []string) map[string]string {
	result := make(map[string]string)
	for _, m := range meta {
		parts := strings.SplitN(m, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			val := strings.TrimSpace(parts[1])
			if key != "" {
				result[key] = val
			}
		}
	}
	return result
}

func ParseMetaInteractive(r io.Reader) (map[string]string, error) {
	meta := make(map[string]string)
	reader := bufio.NewReader(r)

	fmt.Println("Введите метаданные в формате key=value. Для окончания ввода нажмите Enter на пустой строке.")

	for {
		fmt.Print("meta> ")
		line, err := reader.ReadString('\n')

		if errors.Is(err, io.EOF) {
			line = strings.TrimSpace(line)
			if line == "" {
				break
			}

			parts := strings.SplitN(line, "=", 2)
			if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
				return nil, errors.New("ошибка при вводе метаданных: неверный формат при EOF")
			}

			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			if key == "" {
				return nil, errors.New("ошибка при вводе метаданных: ключ не может быть пустым")
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
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
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
