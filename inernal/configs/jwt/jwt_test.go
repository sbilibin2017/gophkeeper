package jwt

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTokenManager_GenerateAndGetUser(t *testing.T) {
	secret := "testsecret"
	lifetime := time.Hour
	manager := NewJWTManager(secret, lifetime)

	username := "testuser"

	// Test Generate
	tokenStr, err := manager.Generate(username)
	assert.NoError(t, err)
	assert.NotEmpty(t, tokenStr)

	// Test GetSubject with valid token
	subject, err := manager.GetSubject(tokenStr)
	assert.NoError(t, err)
	assert.Equal(t, username, subject)
}

func TestTokenManager_GetUser_InvalidToken(t *testing.T) {
	manager := NewJWTManager("secret", time.Hour)

	// Malformed token
	_, err := manager.GetSubject("invalid.token.string")
	assert.Error(t, err)
}

func TestTokenManager_GetUser_ExpiredToken(t *testing.T) {
	manager := NewJWTManager("secret", -1*time.Second)

	username := "expireduser"
	tokenStr, err := manager.Generate(username)
	assert.NoError(t, err)

	// Token should already be expired
	_, err = manager.GetSubject(tokenStr)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "token is expired")
}
