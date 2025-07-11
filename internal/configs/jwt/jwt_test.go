package jwt

import (
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateToken(t *testing.T) {
	secret := "supersecretkey"
	username := "testuser"

	tokenString, err := GenerateToken(username, secret)
	require.NoError(t, err)
	require.NotEmpty(t, tokenString)

	// Проверим, что токен валидный и подписан корректно
	parsedToken, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Проверяем метод подписи
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			t.Fatalf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})
	require.NoError(t, err)
	require.True(t, parsedToken.Valid)

	// Проверим, что claims содержат правильные данные
	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	require.True(t, ok)

	assert.Equal(t, username, claims["sub"])

	// Проверим наличие iat и exp и что exp > iat
	iat, ok := claims["iat"].(float64)
	require.True(t, ok)
	exp, ok := claims["exp"].(float64)
	require.True(t, ok)

	assert.True(t, exp > iat)

	// Проверим, что exp примерно через 24 часа
	expectedExp := int64(iat) + 24*3600
	// Допустимая дельта 5 секунд из-за задержки выполнения
	assert.InDelta(t, expectedExp, int64(exp), 5)
}

func TestGenerateToken_EmptyUsername(t *testing.T) {
	secret := "secret"
	username := ""

	token, err := GenerateToken(username, secret)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	// Можно проверить, что в токене пустой sub
	parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	require.NoError(t, err)
	claims := parsedToken.Claims.(jwt.MapClaims)

	assert.Equal(t, "", claims["sub"])
}

func TestGenerateToken_EmptySecret(t *testing.T) {
	username := "user"
	secret := ""

	// При пустом секрете, SignedString не выдаст ошибку, но подпись будет с пустым ключом
	token, err := GenerateToken(username, secret)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	// Попытка парсить с пустым секретом пройдет
	parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	require.NoError(t, err)
	assert.True(t, parsedToken.Valid)
}
