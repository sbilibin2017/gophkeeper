package models

// TokenPayload представляет полезную нагрузку (claims) JWT токена.
type TokenPayload struct {
	UserID   string `json:"user_id"`
	DeviceID string `json:"device_id"`
}
