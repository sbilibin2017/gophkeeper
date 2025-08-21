package models

import "github.com/golang-jwt/jwt/v4"

// TokenPayload представляет данные, которые будут помещены в JWT.
// Содержит идентификатор пользователя и устройства.
type TokenPayload struct {
	UserID   string `json:"user_id"`   // уникальный идентификатор пользователя
	DeviceID string `json:"device_id"` // уникальный идентификатор устройства пользователя
}

// Claims представляет структуру JWT с включёнными данными TokenPayload
// и стандартными полями JWT (RegisteredClaims), такими как время жизни токена, время выпуска и т.д.
type Claims struct {
	UserID               string `json:"user_id"`   // уникальный идентификатор пользователя
	DeviceID             string `json:"device_id"` // уникальный идентификатор устройства пользователя
	jwt.RegisteredClaims        // Встроенные стандартные поля JWT
}
