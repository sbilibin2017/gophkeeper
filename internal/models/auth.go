package models

// AuthRequest представляет данные запроса для аутентификации пользователя.
type AuthRequest struct {
	Username string `json:"username"` // Имя пользователя.
	Password string `json:"password"` // Пароль пользователя.
}

// AuthResponse содержит ответ с токеном аутентификации.
type AuthResponse struct {
	Token string `json:"token"` // JWT или другой токен для авторизации.
}
