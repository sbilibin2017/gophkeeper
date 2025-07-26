package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

// JWTManager handles JWT generation and validation.
type JWTManager struct {
	secretKey     []byte
	tokenLifetime time.Duration
}

// NewJWTManager creates a new JWTManager with secret and lifetime.
func NewJWTManager(secret string, lifetime time.Duration) *JWTManager {
	return &JWTManager{
		secretKey:     []byte(secret),
		tokenLifetime: lifetime,
	}
}

// Generate creates a signed JWT for a given user.
func (tm *JWTManager) Generate(subject string) (string, error) {
	now := time.Now()
	exp := now.Add(tm.tokenLifetime)

	claims := jwt.RegisteredClaims{
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(exp),
		Subject:   subject,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(tm.secretKey)
}

// GetUsername extracts the username from the Subject field of a valid JWT.
func (tm *JWTManager) GetSubject(tokenStr string) (string, error) {
	parsedToken, err := jwt.ParseWithClaims(tokenStr, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return tm.secretKey, nil
	})
	if err != nil {
		return "", err
	}

	if c, ok := parsedToken.Claims.(*jwt.RegisteredClaims); ok && parsedToken.Valid {
		return c.Subject, nil
	}

	return "", errors.New("invalid token")
}
