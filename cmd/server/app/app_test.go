package app

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

// Test that NewServerCommand builds correctly and runs with a minimal config.
func TestNewServerCommand_BuildAndExecute(t *testing.T) {
	cmd := NewCommand()
	require.IsType(t, &cobra.Command{}, cmd)
	require.Equal(t, "сервер", cmd.Use)

	// run with an invalid address to trigger early error path
	cmd.SetArgs([]string{
		"--address", "!!!bad_address",
	})
	err := cmd.Execute()
	require.Error(t, err, "must fail on invalid address")
	defer os.Remove("gophkeeper.db")
}

// Test runHTTP happy path with short-lived context.
func TestRunHTTP_GracefulShutdown(t *testing.T) {
	tmpDir := t.TempDir()
	dbFile := filepath.Join(tmpDir, "test.db")

	// Context that cancels quickly so server exits.
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	errCh := make(chan error, 1)
	go func() {
		errCh <- runHTTP(
			ctx,
			":0",         // let OS pick free port
			dbFile,       // sqlite file path
			"",           // skip migrations
			"testsecret", // jwt secret
			time.Minute,  // jwt exp
		)
	}()

	select {
	case err := <-errCh:
		require.NoError(t, err)
	case <-time.After(2 * time.Second):
		t.Fatal("runHTTP did not return in time")
	}
}

// Test runHTTP returns error if DB path is invalid.
func TestRunHTTP_DBError(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// Passing invalid DSN to force DB connection error
	err := runHTTP(ctx, ":0", "/no/such/dir/file.db", "", "secret", time.Minute)
	require.Error(t, err)
}

// Optional: Test runHTTP exits when receiving signal
func TestRunHTTP_SignalShutdown(t *testing.T) {
	tmpDir := t.TempDir()
	dbFile := filepath.Join(tmpDir, "test2.db")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan struct{})
	go func() {
		_ = runHTTP(ctx, ":0", dbFile, "", "secret", time.Minute)
		close(done)
	}()

	// Simulate signal by canceling context
	cancel()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("runHTTP did not exit after signal")
	}
}
