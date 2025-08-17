package jwt

import (
	"crypto/rand"
	"crypto/rsa"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJWT_GenerateAndParse(t *testing.T) {
	secret := "supersecretkey"
	duration := time.Minute * 5
	service := New(secret, duration)

	userID := "user123"
	deviceID := "device456"

	token, err := service.Generate(userID, deviceID)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	parsedUserID, parsedDeviceID, err := service.Parse(token)
	require.NoError(t, err)
	assert.Equal(t, userID, parsedUserID)
	assert.Equal(t, deviceID, parsedDeviceID)
}

func TestJWT_InvalidToken(t *testing.T) {
	secret := "supersecretkey"
	service := New(secret, time.Minute)

	// Токен с некорректной подписью
	invalidToken := "invalid.token.string"
	_, _, err := service.Parse(invalidToken)
	assert.Error(t, err)
}

func TestJWT_ExpiredToken(t *testing.T) {
	secret := "supersecretkey"
	service := New(secret, -time.Minute) // токен сразу истёк

	userID := "user123"
	deviceID := "device456"

	token, err := service.Generate(userID, deviceID)
	require.NoError(t, err)

	_, _, err = service.Parse(token)
	assert.Error(t, err)
}

func TestJWT_UnexpectedSigningMethod(t *testing.T) {
	secret := "supersecretkey"
	service := New(secret, time.Minute)

	// создаём токен вручную с методом RSA
	otherToken := jwt.NewWithClaims(jwt.SigningMethodRS256, &claims{
		UserID:   "u",
		DeviceID: "d",
	})

	// создаём случайный приватный ключ RSA, чтобы подписать токен
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	tokenStr, err := otherToken.SignedString(privateKey)
	require.NoError(t, err)

	_, _, err = service.Parse(tokenStr)
	assert.ErrorContains(t, err, "unexpected signing method")
}

func TestJWT_Parse_InvalidClaims(t *testing.T) {
	secret := "supersecretkey"
	service := New(secret, time.Minute)

	// создаём токен с другой структурой claims (не совпадает с нашей claims)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user": "user1",
	})
	tokenStr, err := token.SignedString([]byte(secret))
	require.NoError(t, err)

	// парсим токен через наш сервис
	userID, deviceID, err := service.Parse(tokenStr)

	// Должны попасть в ветку "invalid token"
	assert.Error(t, err)
	assert.EqualError(t, err, "invalid token")
	assert.Empty(t, userID)
	assert.Empty(t, deviceID)
}

func TestJWT_GetFromResponse(t *testing.T) {
	j := New("secret", time.Minute)

	// Создаём http.Response с нужным заголовком
	resp := &http.Response{
		Header: make(http.Header),
	}
	resp.Header.Set("Authorization", "Bearer mytoken123")

	token, err := j.GetFromResponse(resp)
	require.NoError(t, err)
	assert.Equal(t, "mytoken123", token)

	// Проверка ошибки при отсутствии заголовка
	resp2 := &http.Response{
		Header: make(http.Header),
	}
	_, err = j.GetFromResponse(resp2)
	assert.Error(t, err)
	assert.EqualError(t, err, "missing Authorization header in response")

	// Проверка ошибки при неверном формате заголовка
	resp3 := &http.Response{
		Header: make(http.Header),
	}
	resp3.Header.Set("Authorization", "Token abcdef")
	_, err = j.GetFromResponse(resp3)
	assert.Error(t, err)
	assert.EqualError(t, err, "invalid Authorization header format")
}

func TestJWT_SetHeader(t *testing.T) {
	j := New("secret", time.Minute)
	token := "mytoken123"

	// Используем httptest.ResponseRecorder для тестирования
	recorder := httptest.NewRecorder()
	j.SetHeader(recorder, token)

	authHeader := recorder.Header().Get("Authorization")
	assert.Equal(t, "Bearer "+token, authHeader)
}
