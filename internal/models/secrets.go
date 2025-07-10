package models

// LoginPassword хранит данные для логина и пароля с дополнительными метаданными.
type LoginPassword struct {
	Username string            `json:"username"` // Имя пользователя (логин)
	Password string            `json:"password"` // Пароль пользователя
	Meta     map[string]string `json:"meta"`     // Дополнительные метаданные в формате ключ-значение
}

// Text хранит произвольный текст с дополнительными метаданными.
type Text struct {
	Content []byte            `json:"content"` // Текстовые данные (в байтах)
	Meta    map[string]string `json:"meta"`    // Дополнительные метаданные в формате ключ-значение
}

// Binary хранит бинарные данные с дополнительными метаданными.
type Binary struct {
	Content []byte            `json:"content"` // Бинарные данные файла
	Meta    map[string]string `json:"meta"`    // Дополнительные метаданные в формате ключ-значение
}

// Card хранит данные банковской карты с дополнительными метаданными.
type Card struct {
	Number string            `json:"number"` // Номер карты
	Exp    string            `json:"exp"`    // Срок действия карты (например, "12/25")
	CVV    string            `json:"cvv"`    // CVV-код карты
	Meta   map[string]string `json:"meta"`   // Дополнительные метаданные в формате ключ-значение
}
