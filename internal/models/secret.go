package models

import (
	"time"
)

// Константы, определяющие типы сохраняемых секретов.
const (
	LoginPassword = "login_password" // Логин и пароль
	Text          = "text"           // Текстовая информация
	Binary        = "binary"         // Бинарные данные
	Card          = "card"           // Платёжная карта
)

// SecretDB представляет основную метаинформацию о сохранённом секрете.
// Содержит идентификаторы и временные метки, общие для всех типов секретов.
type SecretDB struct {
	SecretID  string    `json:"secret_id" db:"secret_id"`   // Уникальный идентификатор секрета (UUID)
	TypeID    string    `json:"type_id" db:"type_id"`       // Идентификатор типа секрета, ссылается на TypeDB.TypeID
	OwnerID   string    `json:"owner_id" db:"owner_id"`     // Идентификатор владельца секрета (UUID)
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"` // Временная метка последнего обновления
}

// TypeDB представляет определение типа секрета.
type TypeDB struct {
	TypeID    string    `json:"type_id" db:"type_id"`       // Уникальный идентификатор типа секрета
	Name      string    `json:"name" db:"name"`             // Человекочитаемое имя типа секрета
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"` // Временная метка последнего обновления
}

// LoginPasswordDB представляет секрет типа "login_password".
// Сохраняет учетные данные и необязательные метаданные.
type LoginPasswordDB struct {
	SecretID string            `json:"secret_id" db:"secret_id"` // Уникальный идентификатор секрета (UUID)
	Login    string            `json:"login" db:"login"`         // Имя пользователя или логин
	Password string            `json:"password" db:"password"`   // Пароль
	Meta     map[string]string `json:"meta,omitempty" db:"meta"` // Необязательные метаданные
}

// PayloadTextDB представляет секрет типа "text".
// Сохраняет произвольный текст и необязательные метаданные.
type PayloadTextDB struct {
	SecretID string            `json:"secret_id" db:"secret_id"` // Уникальный идентификатор секрета (UUID)
	Content  string            `json:"content" db:"content"`     // Основной текст
	Meta     map[string]string `json:"meta,omitempty" db:"meta"` // Необязательные метаданные
}

// PayloadBinaryDB представляет секрет типа "binary".
// Сохраняет необработанные бинарные данные и необязательные метаданные.
type PayloadBinaryDB struct {
	SecretID string            `json:"secret_id" db:"secret_id"` // Уникальный идентификатор секрета (UUID)
	Data     []byte            `json:"data" db:"data"`           // Сырые бинарные данные
	Meta     map[string]string `json:"meta,omitempty" db:"meta"` // Необязательные метаданные
}

// PayloadCardDB представляет секрет типа "card".
// Сохраняет данные платёжной карты и необязательные метаданные.
type PayloadCardDB struct {
	SecretID string            `json:"secret_id" db:"secret_id"` // Уникальный идентификатор секрета (UUID)
	Number   string            `json:"number" db:"number"`       // Номер карты
	Holder   string            `json:"holder" db:"holder"`       // Имя держателя карты
	ExpMonth int               `json:"exp_month" db:"exp_month"` // Месяц окончания срока действия (1-12)
	ExpYear  int               `json:"exp_year" db:"exp_year"`   // Год окончания срока действия (четыре цифры)
	CVV      string            `json:"cvv" db:"cvv"`             // CVV-код
	Meta     map[string]string `json:"meta,omitempty" db:"meta"` // Необязательные метаданные (например, банк, тип карты)
}
