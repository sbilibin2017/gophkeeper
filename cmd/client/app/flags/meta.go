package flags

import (
	"fmt"
	"strings"
)

// MetaFlag представляет собой пользовательский флаг командной строки,
// который хранит набор ключ-значение в формате map[string]string.
// Используется для передачи метаданных в формате "ключ=значение".
type MetaFlag map[string]string

// String возвращает строковое представление MetaFlag в виде
// ключ-значение, разделённых запятыми, например: "key1=value1, key2=value2".
func (m *MetaFlag) String() string {
	var parts []string
	for k, v := range *m {
		parts = append(parts, fmt.Sprintf("%s=%s", k, v))
	}
	return strings.Join(parts, ", ")
}

// Set парсит строку формата "ключ=значение" и сохраняет пару в MetaFlag.
// Возвращает ошибку, если строка не содержит символа '='
// или если ключ пустой.
func (m *MetaFlag) Set(value string) error {
	parts := strings.SplitN(value, "=", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid meta format: %s (expected key=value)", value)
	}
	key := strings.TrimSpace(parts[0])
	val := strings.TrimSpace(parts[1])
	if key == "" {
		return fmt.Errorf("invalid meta format: key is empty")
	}
	if *m == nil {
		*m = make(map[string]string)
	}
	(*m)[key] = val
	return nil
}

// Type возвращает тип флага, используется для интеграции
// с пакетом flag и похожими механизмами.
func (m *MetaFlag) Type() string {
	return "meta"
}
