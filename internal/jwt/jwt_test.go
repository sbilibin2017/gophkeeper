package jwt

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/stretchr/testify/assert"
)

func TestJWT_GenerateAndGetUsername(t *testing.T) {
	j := New(WithSecret("secret"), WithLifetime(time.Minute))

	tokenStr, err := j.Generate("user123")
	assert.NoError(t, err)
	assert.NotEmpty(t, tokenStr)

	username, err := j.GetUsername(tokenStr)
	assert.NoError(t, err)
	assert.Equal(t, "user123", username)
}

func TestJWT_GetUsername_InvalidToken(t *testing.T) {
	j := New(WithSecret("secret"), WithLifetime(time.Minute))

	_, err := j.GetUsername("invalid.token.string")
	assert.Error(t, err)
}

func TestJWT_GetUsername_ExpiredToken(t *testing.T) {
	j := New(WithSecret("secret"), WithLifetime(-time.Minute)) // expired token

	tokenStr, err := j.Generate("user123")
	assert.NoError(t, err)

	_, err = j.GetUsername(tokenStr)
	assert.Error(t, err)
}

func TestJWT_GetUsername_UnexpectedSigningMethod(t *testing.T) {
	j := New(WithSecret("secret"), WithLifetime(time.Minute))

	// Manually create token with alg=none (unsigned token)
	header := `{"alg":"none","typ":"JWT"}`
	payload := `{"username":"user","iat":1,"exp":1000}`

	encode := func(s string) string {
		return jwt.EncodeSegment([]byte(s))
	}

	tokenStr := encode(header) + "." + encode(payload) + "."

	_, err := j.GetUsername(tokenStr)
	assert.Error(t, err)
	assert.Equal(t, "unexpected signing method", err.Error())
}
