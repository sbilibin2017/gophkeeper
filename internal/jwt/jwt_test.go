package jwt

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/sbilibin2017/gophkeeper/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestJWT_GenerateAndParse(t *testing.T) {
	secret := "testsecret"
	duration := time.Minute * 10
	j := New(secret, duration)

	payload := models.TokenPayload{
		UserID:   "user123",
		DeviceID: "device456",
	}

	tokenString, err := j.Generate(&payload)
	assert.NoError(t, err)
	assert.NotEmpty(t, tokenString)

	claims, err := j.Parse(tokenString)
	assert.NoError(t, err)
	assert.Equal(t, payload.UserID, claims.UserID)
	assert.Equal(t, payload.DeviceID, claims.DeviceID)
}

func TestJWT_GetFromRequest(t *testing.T) {
	j := New("secret", time.Minute*10)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer mytoken123")

	token, err := j.GetFromRequest(req)
	assert.NoError(t, err)
	assert.Equal(t, "mytoken123", token)
}

func TestJWT_GetFromRequest_Invalid(t *testing.T) {
	j := New("secret", time.Minute*10)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "InvalidTokenFormat")

	token, err := j.GetFromRequest(req)
	assert.Error(t, err)
	assert.Empty(t, token)
}

func TestJWT_GetFromResponse(t *testing.T) {
	j := New("secret", time.Minute*10)

	rec := httptest.NewRecorder()
	rec.Header().Set("Authorization", "Bearer mytoken456")

	token, err := j.GetFromResponse(rec.Result())
	assert.NoError(t, err)
	assert.Equal(t, "mytoken456", token)
}

func TestJWT_SetHeader(t *testing.T) {
	j := New("secret", time.Minute*10)

	rec := httptest.NewRecorder()
	j.SetHeader(rec, "newtoken789")

	assert.Equal(t, "Bearer newtoken789", rec.Header().Get("Authorization"))
}
