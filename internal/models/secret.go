package models

import (
	"time"
)

// Константы, определяющие типы хранимых секретов.
const (
	LoginPassword = "login_password" // Логин и пароль
	Text          = "text"           // Текстовая информация
	Binary        = "binary"         // Двоичные данные
	Card          = "card"           // Платёжная карта
)

// Secret представляет собой универсальную структуру хранения секрета любого типа.
type Secret struct {
	ID        string    `json:"id" db:"id"`                 // Уникальный идентификатор секрета (UUID)
	OwnerID   string    `json:"owner_id" db:"owner_id"`     // Идентификатор владельца секрета (UUID)
	SType     string    `json:"type" db:"type"`             // Тип секрета (один из констант выше)
	Payload   []byte    `json:"payload" db:"payload"`       // Содержимое секрета в сериализованном виде (JSON)
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"` // Дата и время последнего обновления
}

// PayloadLoginPassword описывает данные секрета типа "логин и пароль".
type PayloadLoginPassword struct {
	Login    string            `json:"login"`          // Имя пользователя или логин
	Password string            `json:"password"`       // Пароль
	Meta     map[string]string `json:"meta,omitempty"` // Дополнительные метаданные (например, сайт, описание и т.п.)
}

// PayloadText описывает данные секрета типа "текст".
type PayloadText struct {
	Content string            `json:"content"`        // Основной текст
	Meta    map[string]string `json:"meta,omitempty"` // Дополнительные метаданные
}

// PayloadBinary описывает данные секрета типа "двоичные данные".
type PayloadBinary struct {
	Data []byte            `json:"data"`           // Сырые двоичные данные
	Meta map[string]string `json:"meta,omitempty"` // Дополнительные метаданные
}

// PayloadCard описывает данные секрета типа "платёжная карта".
type PayloadCard struct {
	Number   string            `json:"number"`         // Номер карты
	Holder   string            `json:"holder"`         // Имя владельца карты
	ExpMonth int               `json:"exp_month"`      // Месяц окончания срока действия
	ExpYear  int               `json:"exp_year"`       // Год окончания срока действия
	CVV      string            `json:"cvv"`            // CVV-код
	Meta     map[string]string `json:"meta,omitempty"` // Дополнительные метаданные (например, банк, тип карты)
}
