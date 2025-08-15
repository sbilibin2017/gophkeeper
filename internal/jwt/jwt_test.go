package jwt

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

func TestJWT_GenerateAndGetUsername(t *testing.T) {
	tests := []struct {
		name      string
		secret    string
		ttl       time.Duration
		userID    string
		wantError bool
	}{
		{
			name:   "default secret and ttl",
			userID: "user123",
		},
		{
			name:   "custom secret",
			secret: "mysecret",
			userID: "user456",
		},
		{
			name:   "short ttl",
			ttl:    time.Second,
			userID: "user789",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := []Opt{}
			if tt.secret != "" {
				opts = append(opts, WithSecret(tt.secret))
			}
			if tt.ttl != 0 {
				opts = append(opts, WithTTL(tt.ttl))
			}

			j := New(opts...)
			token, err := j.Generate(tt.userID)
			assert.NoError(t, err)
			assert.NotEmpty(t, token)

			gotUserID, err := j.GetUsername(token)
			assert.NoError(t, err)
			assert.Equal(t, tt.userID, gotUserID)
		})
	}
}

func TestJWT_GetUsername_InvalidToken(t *testing.T) {
	j := New(WithSecret("secret"))

	tests := []struct {
		name      string
		token     string
		wantError bool
	}{
		{
			name:      "empty token",
			token:     "",
			wantError: true,
		},
		{
			name:      "random string",
			token:     "abc.def.ghi",
			wantError: true,
		},
		{
			name:      "wrong secret",
			token:     func() string { tok, _ := New(WithSecret("other")).Generate("u1"); return tok }(),
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userID, err := j.GetUsername(tt.token)
			if tt.wantError {
				assert.Error(t, err)
				assert.Empty(t, userID)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, userID)
			}
		})
	}
}

func TestJWT_GetUsername_InvalidSigningMethod(t *testing.T) {
	j := New(WithSecret("secret"))

	// Создаём токен с методом None
	token := jwt.NewWithClaims(jwt.SigningMethodNone, &claims{UserID: "123"})
	tokenStr, err := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
	assert.NoError(t, err)

	userID, err := j.GetUsername(tokenStr)
	assert.Empty(t, userID)
	assert.Error(t, err)
}
