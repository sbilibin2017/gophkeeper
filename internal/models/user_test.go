package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestUser_GetSecretName(t *testing.T) {
	user := &User{
		SecretName: "user-secret",
	}
	assert.Equal(t, "user-secret", user.GetSecretName(), "GetSecretName should return SecretName field")
}

func TestUser_GetUpdatedAt(t *testing.T) {
	now := time.Now()
	user := &User{
		UpdatedAt: now,
	}
	assert.Equal(t, now, user.GetUpdatedAt(), "GetUpdatedAt should return UpdatedAt field")
}

func TestUserData_Fields(t *testing.T) {
	meta := "user metadata"
	userData := &UserPayload{
		Login:    "testuser",
		Password: "securepassword",
		Meta:     &meta,
	}

	assert.Equal(t, "testuser", userData.Login)
	assert.Equal(t, "securepassword", userData.Password)
	assert.NotNil(t, userData.Meta)
	assert.Equal(t, "user metadata", *userData.Meta)
}
