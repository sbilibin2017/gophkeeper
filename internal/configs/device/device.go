package device

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

// GetDeviceID возвращает уникальный идентификатор устройства.
// Если идентификатор ранее был сохранён в файле, он читается.
// Если файла нет — создается новый UUID и сохраняется.
func GetDeviceID() (string, error) {
	// файл для хранения ID (в текущей директории)
	filePath := filepath.Join(".", ".device_id")

	// Проверяем, есть ли сохранённый ID
	if data, err := os.ReadFile(filePath); err == nil {
		id := string(data)
		if id != "" {
			return id, nil
		}
	}

	// Генерируем новый UUID
	newID := uuid.New().String()

	// Сохраняем его в файл
	if err := os.WriteFile(filePath, []byte(newID), 0644); err != nil {
		return "", errors.New("не удалось сохранить deviceID: " + err.Error())
	}

	return newID, nil
}
