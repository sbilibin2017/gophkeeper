package jwt

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/sbilibin2017/gophkeeper/internal/models"
)

// JWT предоставляет методы для генерации и парсинга JWT-токенов.
type JWT struct {
	secretKey     []byte
	tokenDuration time.Duration
}

func New(secret string, duration time.Duration) *JWT {
	return &JWT{
		secretKey:     []byte(secret),
		tokenDuration: duration,
	}
}

// Generate создаёт JWT на основе TokenRequest
func (j *JWT) Generate(payload *models.TokenPayload) (string, error) {
	c := models.Claims{
		UserID:   payload.UserID,
		DeviceID: payload.DeviceID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.tokenDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, c)
	return token.SignedString(j.secretKey)
}

// Parse извлекает Claims из JWT и возвращает *models.Claims
func (j *JWT) Parse(tokenString string) (*models.Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &models.Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return j.secretKey, nil
	})
	if err != nil {
		return nil, err
	}

	c, ok := token.Claims.(*models.Claims)
	if !ok || !token.Valid || c.UserID == "" || c.DeviceID == "" {
		return nil, errors.New("invalid token")
	}

	return c, nil
}

// GetFromRequest извлекает JWT-токен из заголовка Authorization
func (j *JWT) GetFromRequest(req *http.Request) (string, error) {
	authHeader := req.Header.Get("Authorization")
	if authHeader == "" {
		return "", errors.New("missing Authorization header")
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" || parts[1] == "" {
		return "", errors.New("invalid Authorization header format")
	}

	return parts[1], nil
}

// GetFromResponse извлекает JWT-токен из заголовка Authorization
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
