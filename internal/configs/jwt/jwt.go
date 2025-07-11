package jwt

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// GenerateToken генерирует JWT-токен с заданным логином и временем жизни
func GenerateToken(username string, secret string) (string, error) {
	claims := jwt.MapClaims{
		"sub": username,
		"exp": time.Now().Add(time.Hour * 24).Unix(),
		"iat": time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}
