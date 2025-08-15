package jwt

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/golang-jwt/jwt/v5"
)

// JWT структура для генерации и валидации токенов
type JWT struct {
	secret string
	ttl    time.Duration
}

// claims — приватная структура для JWT токена
type claims struct {
	UserID string `json:"user_id"` // UUID пользователя
	jwt.RegisteredClaims
}

// Opt тип для функциональных опций
type Opt func(*JWT)

// WithSecret задает секрет для JWT
func WithSecret(secret string) Opt {
	return func(j *JWT) {
		j.secret = secret
	}
}

// WithTTL задает время жизни токена
func WithTTL(ttl time.Duration) Opt {
	return func(j *JWT) {
		j.ttl = ttl
	}
}

// New создает JWT с опциями
func New(opts ...Opt) *JWT {
	j := &JWT{
		secret: "secret",
		ttl:    time.Hour,
	}

	for _, opt := range opts {
		opt(j)
	}

	return j
}

// Generate создает JWT токен для заданного userID
func (j *JWT) Generate(userID string) (string, error) {
	c := &claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.ttl)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, c)
	return token.SignedString([]byte(j.secret))
}

// GetUsername извлекает userID из токена
func (j *JWT) GetUsername(tokenStr string) (string, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(j.secret), nil
	})
	if err != nil {
		return "", err
	}

	if c, ok := token.Claims.(*claims); ok && token.Valid {
		return c.UserID, nil
	}

	return "", errors.New("invalid token")
}

// GetTokenFromRestyResponse извлекает JWT токен из заголовка Authorization HTTP-ответа Resty.
func GetTokenFromRestyResponse(resp *resty.Response) (string, error) {
	authHeader := resp.Header().Get("Authorization")
	if authHeader == "" {
		return "", fmt.Errorf("missing Authorization header")
	}

	const prefix = "Bearer "
	if !strings.HasPrefix(authHeader, prefix) {
		return "", fmt.Errorf("invalid Authorization header format")
	}

	token := strings.TrimPrefix(authHeader, prefix)
	if token == "" {
		return "", fmt.Errorf("empty token in Authorization header")
	}

	return token, nil
}
