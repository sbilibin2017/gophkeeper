package jwt

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/metadata"
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

// --- New tests for GetTokenFromHeader ---

func TestGetTokenFromHeader_Valid(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer sometoken123")

	token, err := GetTokenFromHeader(req)
	assert.NoError(t, err)
	assert.Equal(t, "sometoken123", token)
}

func TestGetTokenFromHeader_MissingHeader(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)

	token, err := GetTokenFromHeader(req)
	assert.Error(t, err)
	assert.Empty(t, token)
	assert.Equal(t, "authorization header missing", err.Error())
}

func TestGetTokenFromHeader_InvalidFormat(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "InvalidFormatToken")

	token, err := GetTokenFromHeader(req)
	assert.Error(t, err)
	assert.Empty(t, token)
	assert.Equal(t, "invalid authorization header format", err.Error())
}

func TestGetTokenFromHeader_WrongScheme(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Basic sometoken")

	token, err := GetTokenFromHeader(req)
	assert.Error(t, err)
	assert.Empty(t, token)
	assert.Equal(t, "invalid authorization header format", err.Error())
}

// --- New tests for GetTokenFromContext ---

func TestGetTokenFromContext_Valid(t *testing.T) {
	md := metadata.New(map[string]string{
		"authorization": "Bearer grpc_token_456",
	})
	ctx := metadata.NewIncomingContext(context.Background(), md)

	token, err := GetTokenFromContext(ctx)
	assert.NoError(t, err)
	assert.Equal(t, "grpc_token_456", token)
}

func TestGetTokenFromContext_MissingMetadata(t *testing.T) {
	ctx := context.Background()

	token, err := GetTokenFromContext(ctx)
	assert.Error(t, err)
	assert.Empty(t, token)
	assert.Equal(t, "missing metadata in context", err.Error())
}

func TestGetTokenFromContext_MissingAuthorization(t *testing.T) {
	md := metadata.New(nil) // empty metadata
	ctx := metadata.NewIncomingContext(context.Background(), md)

	token, err := GetTokenFromContext(ctx)
	assert.Error(t, err)
	assert.Empty(t, token)
	assert.Equal(t, "authorization metadata missing", err.Error())
}

func TestGetTokenFromContext_InvalidFormat(t *testing.T) {
	md := metadata.New(map[string]string{
		"authorization": "InvalidTokenFormat",
	})
	ctx := metadata.NewIncomingContext(context.Background(), md)

	token, err := GetTokenFromContext(ctx)
	assert.Error(t, err)
	assert.Empty(t, token)
	assert.Equal(t, "invalid authorization metadata format", err.Error())
}

func TestGetTokenFromContext_WrongScheme(t *testing.T) {
	md := metadata.New(map[string]string{
		"authorization": "Basic grpc_token",
	})
	ctx := metadata.NewIncomingContext(context.Background(), md)

	token, err := GetTokenFromContext(ctx)
	assert.Error(t, err)
	assert.Empty(t, token)
	assert.Equal(t, "invalid authorization metadata format", err.Error())
}
