package jwt

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJWT_GenerateAndParse(t *testing.T) {
	secret := "mysecret"
	username := "testuser"
	j := New(WithSecret(secret), WithLifetime(time.Minute))

	token, err := j.Generate(username)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	parsedUsername, err := j.Parse(token)
	require.NoError(t, err)
	assert.Equal(t, username, parsedUsername)
}

func TestJWT_Parse_ExpiredToken(t *testing.T) {
	secret := "mysecret"
	username := "testuser"
	j := New(WithSecret(secret), WithLifetime(-time.Minute)) // already expired

	token, err := j.Generate(username)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	parsedUsername, err := j.Parse(token)
	assert.Error(t, err)
	assert.Empty(t, parsedUsername)
}

func TestJWT_Parse_InvalidToken(t *testing.T) {
	j := New(WithSecret("secret"))
	username, err := j.Parse("invalid.token.value")
	assert.Error(t, err)
	assert.Empty(t, username)
}

func TestJWT_Parse_InvalidSignature(t *testing.T) {
	j1 := New(WithSecret("secret1"), WithLifetime(time.Minute))
	j2 := New(WithSecret("secret2")) // different secret

	token, err := j1.Generate("user")
	require.NoError(t, err)

	username, err := j2.Parse(token)
	assert.Error(t, err)
	assert.Empty(t, username)
}

func TestJWT_Parse_InvalidSigningMethod(t *testing.T) {
	// Generate RSA key for RS256 signing
	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	// Create token signed with RS256
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, &claims{
		Username: "user",
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute)),
		},
	})

	tokenStr, err := token.SignedString(privKey)
	require.NoError(t, err)

	// Create JWT instance expecting HS256 tokens
	j := New(WithSecret("secret"))

	username, err := j.Parse(tokenStr)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unexpected signing method")
	assert.Equal(t, "", username)
}
