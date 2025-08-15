package app

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIntegration_RegisterCommand(t *testing.T) {
	// Create temporary migrations directory with a valid goose migration
	migrationsDir := "migrations"
	require.NoError(t, os.Mkdir(migrationsDir, 0755))
	defer os.RemoveAll(migrationsDir)

	migrationFile := filepath.Join(migrationsDir, "0001_init.sql")
	migrationContent := `-- +goose Up
CREATE TABLE dummy(id INTEGER);

-- +goose Down
DROP TABLE dummy;
`
	require.NoError(t, os.WriteFile(migrationFile, []byte(migrationContent), 0644))

	// Clean up device ID and DB
	deviceFile := ".device_id"
	defer os.Remove(deviceFile)
	dbFile := "testuser.db"
	defer os.Remove(dbFile)

	// Start a test HTTP server returning plain text
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && r.URL.Path == "/register" {
			w.Header().Set("Authorization", "Bearer integration-token")
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("integration-key"))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	// Prepare CLI command
	cmd := NewRegisterCommand()
	var output bytes.Buffer
	cmd.SetOut(&output)
	cmd.SetArgs([]string{
		"--username", "testuser",
		"--password", "testpass",
		"--server-url", ts.URL,
	})

	// Run the command (fully integration)
	err := cmd.Execute()
	require.NoError(t, err)

	out := output.String()
	require.Contains(t, out, "Регистрация успешна")
	require.Contains(t, out, "Приватный ключ: integration-key")
	require.Contains(t, out, "Токен: integration-token")

	// Check device ID file exists
	data, err := os.ReadFile(deviceFile)
	require.NoError(t, err)
	require.NotEmpty(t, data)

	// Check DB file exists
	_, err = os.Stat(dbFile)
	require.NoError(t, err)
}

func TestNewRootCommandStructure(t *testing.T) {
	root := NewRootCommand()
	require.NotNil(t, root)
	require.Equal(t, "gophkeeper-client", root.Use)
	require.Len(t, root.Commands(), 0)
}

func TestIntegration_DeviceIDAndServer(t *testing.T) {
	// Create temporary migrations directory
	migrationsDir, err := os.MkdirTemp("", "migrations")
	require.NoError(t, err)
	defer os.RemoveAll(migrationsDir)

	migrationFile := filepath.Join(migrationsDir, "0001_init.sql")
	migrationContent := `-- +goose Up
CREATE TABLE IF NOT EXISTS dummy(id INTEGER);

-- +goose Down
DROP TABLE IF EXISTS dummy;
`
	err = os.WriteFile(migrationFile, []byte(migrationContent), 0644)
	require.NoError(t, err)

	// Remove any existing device ID to simulate fresh run
	deviceFile := ".device_id"
	os.Remove(deviceFile)
	defer os.Remove(deviceFile)

	// Remove any existing test DB
	dbFile := "testuser.db"
	os.Remove(dbFile)
	defer os.Remove(dbFile)

	// Start a test HTTP server returning plain text
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && r.URL.Path == "/register" {
			w.Header().Set("Authorization", "Bearer integration-token")
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusOK)
			// Return private key in plain text (as your code expects)
			w.Write([]byte("integration-key"))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	// Get device ID
	deviceID, err := getDeviceID()
	require.NoError(t, err)
	require.NotEmpty(t, deviceID)

	// Run register HTTP against test server
	privKey, token, err := runRegisterHTTP(
		context.Background(),
		ts.URL,
		migrationsDir, // point to temporary migrations folder
		"testuser",
		"testpass",
		deviceID,
	)
	require.NoError(t, err)
	require.Equal(t, "integration-token", token)
	require.Equal(t, "integration-key", string(privKey))

	// Ensure device ID file was written
	data, err := os.ReadFile(deviceFile)
	require.NoError(t, err)
	require.Equal(t, deviceID, string(data))

	// Ensure database file exists
	_, err = os.Stat(dbFile)
	require.NoError(t, err)
}
