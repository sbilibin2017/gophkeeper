package server

import (
	"context"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testApiVersion      = "/api/v1"
	testJwtSecretKey    = "testsecret"
	testJwtExp          = time.Hour
	testDatabaseDriver  = "sqlite"
	testDatabaseDSN     = "file::memory:?cache=shared"
	testServerAddrHTTP  = "127.0.0.1:8081"
	testServerAddrGRPC  = "127.0.0.1:50051"
	testMaxOpenConns    = 1
	testMaxIdleConns    = 1
	testConnMaxLifetime = time.Minute
)

// createTempMigrationsDir creates a temp directory with a goose migration file that creates/drops users table
func createTempMigrationsDir(t *testing.T) string {
	tmpDir := t.TempDir()
	migrationFile := filepath.Join(tmpDir, "00001_create_users_table.sql")

	sql := `-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS users (
    username TEXT PRIMARY KEY,
    password_hash TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS users;
-- +goose StatementEnd
`

	err := os.WriteFile(migrationFile, []byte(sql), 0644)
	require.NoError(t, err, "failed to create migration file")

	return tmpDir
}

func TestRunHTTP_StartStop(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	migrationsDir := createTempMigrationsDir(t)

	go func() {
		time.Sleep(200 * time.Millisecond)
		cancel()
	}()

	err := RunHTTP(
		ctx,
		testApiVersion,
		testDatabaseDriver,
		testDatabaseDSN,
		testMaxOpenConns,
		testMaxIdleConns,
		testConnMaxLifetime,
		testJwtSecretKey,
		testJwtExp,
		migrationsDir,
		testServerAddrHTTP,
	)
	assert.NoError(t, err)
}

func TestRunHTTP_RegisterEndpoint(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	migrationsDir := createTempMigrationsDir(t)

	go func() {
		time.Sleep(500 * time.Millisecond)
		cancel()
	}()

	// Start server in goroutine
	go func() {
		_ = RunHTTP(
			ctx,
			testApiVersion,
			testDatabaseDriver,
			testDatabaseDSN,
			testMaxOpenConns,
			testMaxIdleConns,
			testConnMaxLifetime,
			testJwtSecretKey,
			testJwtExp,
			migrationsDir,
			testServerAddrHTTP,
		)
	}()

	// Wait for server to start accepting connections
	waitForHTTPServer(t, testServerAddrHTTP, 5*time.Second)

	// Test /register endpoint with empty payload (should return 400 Bad Request)
	resp, err := http.Post("http://"+testServerAddrHTTP+testApiVersion+"/register", "application/json", nil)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func waitForHTTPServer(t *testing.T, addr string, timeout time.Duration) {
	deadline := time.Now().Add(timeout)
	for {
		conn, err := net.DialTimeout("tcp", addr, time.Millisecond*100)
		if err == nil {
			_ = conn.Close()
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("server did not start within %s", timeout)
		}
		time.Sleep(50 * time.Millisecond)
	}
}

func TestRunGRPC_StartStop(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	migrationsDir := createTempMigrationsDir(t)

	go func() {
		time.Sleep(200 * time.Millisecond)
		cancel()
	}()

	err := RunGRPC(
		ctx,
		testDatabaseDriver,
		testDatabaseDSN,
		testMaxOpenConns,
		testMaxIdleConns,
		testConnMaxLifetime,
		testJwtSecretKey,
		testJwtExp,
		migrationsDir,
		testServerAddrGRPC,
	)
	assert.NoError(t, err)
}

func TestRunHTTP_ListenAndServeError(t *testing.T) {
	migrationsDir := createTempMigrationsDir(t)

	// Bind the address before starting the server to cause address in use error
	ln, err := net.Listen("tcp", testServerAddrHTTP)
	require.NoError(t, err)
	defer ln.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err = RunHTTP(
		ctx,
		testApiVersion,
		testDatabaseDriver,
		testDatabaseDSN,
		testMaxOpenConns,
		testMaxIdleConns,
		testConnMaxLifetime,
		testJwtSecretKey,
		testJwtExp,
		migrationsDir,
		testServerAddrHTTP,
	)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "address already in use")
}

func TestRunGRPC_ServeError(t *testing.T) {
	migrationsDir := createTempMigrationsDir(t)

	// Bind the address before starting the grpc server to cause error
	ln, err := net.Listen("tcp", testServerAddrGRPC)
	require.NoError(t, err)
	defer ln.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err = RunGRPC(
		ctx,
		testDatabaseDriver,
		testDatabaseDSN,
		testMaxOpenConns,
		testMaxIdleConns,
		testConnMaxLifetime,
		testJwtSecretKey,
		testJwtExp,
		migrationsDir,
		testServerAddrGRPC,
	)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "address already in use")
}
