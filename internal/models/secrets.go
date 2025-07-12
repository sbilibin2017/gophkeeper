package models

// UsernamePassword хранит пару логин и пароль.
type UsernamePassword struct {
	Username string            `json:"username"`
	Password string            `json:"password"`
	Meta     map[string]string `json:"meta,omitempty"`
}

// Text хранит произвольные текстовые данные.
type Text struct {
	Content string            `json:"content"`
	Meta    map[string]string `json:"meta,omitempty"`
}

// Binary хранит произвольные бинарные данные.
type Binary struct {
	Data []byte            `json:"data"`
	Meta map[string]string `json:"meta,omitempty"`
}

// BankCard хранит минимально необходимую информацию о банковской карте.
type BankCard struct {
	Number string            `json:"number"`         // Номер карты
	Owner  string            `json:"owner"`          // Владелец карты
	Expiry string            `json:"expiry"`         // Срок действия (MM/YY)
	CVV    string            `json:"cvv"`            // CVV-код
	Meta   map[string]string `json:"meta,omitempty"` // Дополнительная метаинформация
}
