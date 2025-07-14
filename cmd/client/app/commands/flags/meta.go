package flags

import (
	"encoding/json"
)

// PrepareMetaJSON парсит JSON-строку meta и возвращает её обратно в *string.
// Проверяет корректность JSON, возвращает ошибку, если невалидно.
func PrepareMetaJSON(meta string) (*string, error) {
	if meta == "" {
		return nil, nil
	}

	var parsed map[string]string
	if err := json.Unmarshal([]byte(meta), &parsed); err != nil {
		return nil, err
	}

	b, err := json.Marshal(parsed)
	if err != nil {
		return nil, err
	}

	s := string(b)
	return &s, nil
}
