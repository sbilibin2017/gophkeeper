package server

import (
	"testing"
	"time"

	"github.com/sbilibin2017/gophkeeper/internal/address"

	"github.com/stretchr/testify/assert"
)

func TestNewCommand(t *testing.T) {
	cmd := NewCommand()
	assert.NotNil(t, cmd, "NewCommand should return a non-nil command")

	// Test default values
	serverURLFlag := cmd.Flags().Lookup("server-url")
	assert.NotNil(t, serverURLFlag)
	assert.Equal(t, "http://localhost:8080", serverURLFlag.DefValue)

	jwtExpFlag := cmd.Flags().Lookup("jwt-exp")
	assert.NotNil(t, jwtExpFlag)
	assert.Equal(t, time.Hour.String(), jwtExpFlag.DefValue)

	// Test flag parsing by setting custom values
	err := cmd.ParseFlags([]string{
		"--server-url", "grpc://localhost:50051",
		"--database-dsn", "file:test.db",
		"--jwt-secret", "supersecret",
		"--jwt-exp", "2h",
	})
	assert.NoError(t, err)

	// Verify parsed flag values
	serverURL, err := cmd.Flags().GetString("server-url")
	assert.NoError(t, err)
	assert.Equal(t, "grpc://localhost:50051", serverURL)

	databaseDSN, err := cmd.Flags().GetString("database-dsn")
	assert.NoError(t, err)
	assert.Equal(t, "file:test.db", databaseDSN)

	jwtSecret, err := cmd.Flags().GetString("jwt-secret")
	assert.NoError(t, err)
	assert.Equal(t, "supersecret", jwtSecret)

	jwtExp, err := cmd.Flags().GetDuration("jwt-exp")
	assert.NoError(t, err)
	assert.Equal(t, 2*time.Hour, jwtExp)

	// Test address parsing helper
	addr := address.New(serverURL)
	assert.Equal(t, "grpc", addr.Scheme)
	assert.Equal(t, "localhost:50051", addr.Address)
}
