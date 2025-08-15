package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/sbilibin2017/gophkeeper/internal/models"
)

type JWT struct {
	secretKey     []byte
	tokenDuration time.Duration
}

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
func (j *JWT) Generate(payload *models.TokenPayload) (string, error) {
	c := claims{
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

// Parse извлекает Claims из JWT и возвращает TokenPayload
func (j *JWT) Parse(tokenString string) (*models.TokenPayload, error) {
	token, err := jwt.ParseWithClaims(tokenString, &claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return j.secretKey, nil
	})
	if err != nil {
		return nil, err
	}

	if c, ok := token.Claims.(*claims); ok && token.Valid {
		return &models.TokenPayload{
			UserID:   c.UserID,
			DeviceID: c.DeviceID,
		}, nil
	}

	return nil, errors.New("invalid token")
}
