package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewCredentials_WithOptions(t *testing.T) {
	creds := NewCredentials(
		WithUsername("testuser"),
		WithPassword("secret123"),
	)

	assert.Equal(t, "testuser", creds.Username)
	assert.Equal(t, "secret123", creds.Password)
}

func TestNewCredentials_Empty(t *testing.T) {
	creds := NewCredentials()

	assert.Empty(t, creds.Username)
	assert.Empty(t, creds.Password)
}

func TestWithUsername(t *testing.T) {
	creds := &Credentials{}
	opt := WithUsername("alice")
	opt(creds)

	assert.Equal(t, "alice", creds.Username)
}

func TestWithPassword(t *testing.T) {
	creds := &Credentials{}
	opt := WithPassword("mypassword")
	opt(creds)

	assert.Equal(t, "mypassword", creds.Password)
}

func TestNewSecretAddRequest_WithOptions(t *testing.T) {
	req, err := NewSecretAddRequest(
		WithServerURL("https://example.com"),
		WithSType("login"),
		WithFile("/path/to/file"),
		WithInteractive(false),
		WithHMACKey("hmac123"),
		WithRSAPublicKeyPath("/path/to/rsa.pub"),
	)
	assert.NoError(t, err)
	assert.Equal(t, "https://example.com", req.ServerURL)
	assert.Equal(t, "login", req.SType)
	assert.Equal(t, "/path/to/file", req.File)
	assert.False(t, req.Interactive)
	assert.Equal(t, "hmac123", req.HMACKey)
	assert.Equal(t, "/path/to/rsa.pub", req.RSAPublicKeyPath)
}

func TestNewSecretAddRequest_Validation(t *testing.T) {
	// Valid: file specified, interactive false
	req, err := NewSecretAddRequest(WithFile("/tmp/file"))
	assert.NoError(t, err)
	assert.Equal(t, "/tmp/file", req.File)

	// Valid: file empty, interactive true
	req, err = NewSecretAddRequest(WithInteractive(true))
	assert.NoError(t, err)
	assert.True(t, req.Interactive)

	// Invalid: neither file nor interactive specified
	req, err = NewSecretAddRequest()
	assert.Error(t, err)
	assert.Nil(t, req)
	assert.Equal(t, "either file or interactive must be specified", err.Error())

	// Invalid: both file and interactive specified
	req, err = NewSecretAddRequest(WithFile("/tmp/file"), WithInteractive(true))
	assert.Error(t, err)
	assert.Nil(t, req)
	assert.Equal(t, "file and interactive cannot be used together", err.Error())
}
