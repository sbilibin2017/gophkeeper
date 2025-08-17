package jwt

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

// JWT предоставляет методы для генерации и парсинга JWT-токенов.
type JWT struct {
	secretKey     []byte
	tokenDuration time.Duration
}

// приватная структура для хранения claims
type claims struct {
	UserID   string `json:"user_id"`
	DeviceID string `json:"device_id"`
	jwt.RegisteredClaims
}

// New создаёт сервис с ключом и временем жизни токена
func New(secret string, duration time.Duration) *JWT {
	return &JWT{
		secretKey:     []byte(secret),
		tokenDuration: duration,
	}
}

// Generate создаёт JWT с user_id и device_id
func (j *JWT) Generate(userID string, deviceID string) (tokenString string, err error) {
	c := claims{
		UserID:   userID,
		DeviceID: deviceID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.tokenDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, c)
	return token.SignedString(j.secretKey)
}

// Parse извлекает Claims из JWT и возвращает userID и deviceID
func (j *JWT) Parse(tokenString string) (userID string, deviceID string, err error) {
	token, err := jwt.ParseWithClaims(tokenString, &claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return j.secretKey, nil
	})
	if err != nil {
		return
	}

	c, ok := token.Claims.(*claims)
	if !ok || !token.Valid || c.UserID == "" || c.DeviceID == "" {
		err = errors.New("invalid token")
		return
	}

	userID = c.UserID
	deviceID = c.DeviceID
	return
}

// GetFromResponse извлекает JWT-токен из заголовка Authorization в формате Bearer.
// Возвращает токен или ошибку, если заголовок отсутствует или имеет неправильный формат.
func (j *JWT) GetFromResponse(resp *http.Response) (string, error) {
	authHeader := resp.Header.Get("Authorization")
	if authHeader == "" {
		return "", errors.New("missing Authorization header in response")
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return "", errors.New("invalid Authorization header format")
	}

	return parts[1], nil
}

// SetHeader устанавливает JWT-токен в заголовок Authorization HTTP-ответа
func (j *JWT) SetHeader(w http.ResponseWriter, token string) {
	w.Header().Set("Authorization", "Bearer "+token)
}
