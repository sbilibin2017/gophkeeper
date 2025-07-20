package client

import (
	"bytes"
	"context"
	"encoding/json"
	"net"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/sbilibin2017/gophkeeper/internal/models"
	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc/auth"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

// --- HTTP Test Server for Login ---

func startLoginTestHTTPServer(t *testing.T) (*http.Server, string) {
	handler := http.NewServeMux()
	handler.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		resp := models.AuthResponse{Token: "logintoken123"}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	srv := &http.Server{Handler: handler}

	ln, err := net.Listen("tcp", "127.0.0.1:0") // random port
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}

	go func() {
		if err := srv.Serve(ln); err != nil && err != http.ErrServerClosed {
			t.Logf("HTTP server error: %v", err)
		}
	}()

	return srv, "http://" + ln.Addr().String()
}

// --- gRPC Test Server for Login ---

type minimalLoginAuthServer struct {
	pb.UnimplementedAuthServiceServer
}

func (s *minimalLoginAuthServer) Login(ctx context.Context, req *pb.AuthRequest) (*pb.AuthResponse, error) {
	return &pb.AuthResponse{Token: "logintoken123"}, nil
}

func startLoginTestGRPCServer(t *testing.T) (*grpc.Server, string) {
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}

	grpcSrv := grpc.NewServer()
	pb.RegisterAuthServiceServer(grpcSrv, &minimalLoginAuthServer{})

	go func() {
		if err := grpcSrv.Serve(lis); err != nil {
			t.Logf("gRPC server error: %v", err)
		}
	}()

	return grpcSrv, lis.Addr().String()
}

// --- Cleanup helper ---

func cleanupLoginTestDBFile(t *testing.T) {
	t.Helper()
	t.Cleanup(func() {
		err := os.Remove("client.db")
		if err != nil && !os.IsNotExist(err) {
			t.Errorf("failed to remove client.db: %v", err)
		}
	})
}

// --- Tests ---

func TestLoginHTTP(t *testing.T) {
	cleanupLoginTestDBFile(t)

	srv, url := startLoginTestHTTPServer(t)
	defer srv.Shutdown(context.Background())

	time.Sleep(100 * time.Millisecond) // wait for server

	resp, err := loginHTTP(context.Background(), "user", "pass", url, "", "")
	require.NoError(t, err)
	require.Equal(t, "logintoken123", resp.Token)
}

func TestLoginGRPC(t *testing.T) {
	cleanupLoginTestDBFile(t)

	grpcSrv, addr := startLoginTestGRPCServer(t)
	defer grpcSrv.GracefulStop()

	time.Sleep(100 * time.Millisecond) // wait for server

	url := "grpc://" + addr

	resp, err := loginGRPC(context.Background(), "user", "pass", url, "", "")
	require.NoError(t, err)
	require.Equal(t, "logintoken123", resp.Token)
}

func TestRegisterLoginCommand(t *testing.T) {
	// Backup original funcs
	origLoginHTTPFunc := loginHTTPFunc
	origLoginGRPCFunc := loginGRPCFunc
	defer func() {
		loginHTTPFunc = origLoginHTTPFunc
		loginGRPCFunc = origLoginGRPCFunc
	}()

	// Mock login funcs
	loginHTTPFunc = func(ctx context.Context, username, password, authURL, tlsCertFile, tlsKeyFile string) (*models.AuthResponse, error) {
		return &models.AuthResponse{Token: "http-login-token"}, nil
	}
	loginGRPCFunc = func(ctx context.Context, username, password, authURL, tlsCertFile, tlsKeyFile string) (*models.AuthResponse, error) {
		return &models.AuthResponse{Token: "grpc-login-token"}, nil
	}

	root := &cobra.Command{Use: "root"}
	RegisterLoginCommand(root)

	// Test HTTP login
	root.SetArgs([]string{
		"login",
		"--username", "alice",
		"--password", "pass",
		"--auth-url", "https://example.com",
		"--tls-client-cert", "cert.pem",
		"--tls-client-key", "key.pem",
	})

	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)

	err := root.Execute()
	require.NoError(t, err)

	output := buf.String()
	require.Contains(t, output, "http-login-token")

	// Test gRPC login
	root.SetArgs([]string{
		"login",
		"--username", "bob",
		"--password", "pass",
		"--auth-url", "grpc://example.com",
		"--tls-client-cert", "cert.pem",
		"--tls-client-key", "key.pem",
	})

	buf.Reset()

	err = root.Execute()
	require.NoError(t, err)

	output = buf.String()
	require.Contains(t, output, "grpc-login-token")
}
