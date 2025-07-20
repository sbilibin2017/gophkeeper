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

// --- HTTP Test Server ---

func startRegisterTestHTTPServer(t *testing.T) (*http.Server, string) {
	handler := http.NewServeMux()
	handler.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		resp := models.AuthResponse{Token: "testtoken123"}
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

// --- gRPC Test Server ---

type minimalRegisterAuthServer struct {
	pb.UnimplementedAuthServiceServer
}

func (s *minimalRegisterAuthServer) Register(ctx context.Context, req *pb.AuthRequest) (*pb.AuthResponse, error) {
	return &pb.AuthResponse{Token: "testtoken123"}, nil
}

func startRegisterTestGRPCServer(t *testing.T) (*grpc.Server, string) {
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}

	grpcSrv := grpc.NewServer()
	pb.RegisterAuthServiceServer(grpcSrv, &minimalRegisterAuthServer{})

	go func() {
		if err := grpcSrv.Serve(lis); err != nil {
			t.Logf("gRPC server error: %v", err)
		}
	}()

	return grpcSrv, lis.Addr().String()
}

// --- Tests ---

func cleanupRegisterTestDBFile(t *testing.T) {
	t.Helper()
	t.Cleanup(func() {
		err := os.Remove("client.db")
		if err != nil && !os.IsNotExist(err) {
			t.Errorf("failed to remove client.db: %v", err)
		}
	})
}

func TestRegisterHTTP(t *testing.T) {
	cleanupRegisterTestDBFile(t)

	srv, url := startRegisterTestHTTPServer(t)
	defer srv.Shutdown(context.Background())

	time.Sleep(100 * time.Millisecond)

	resp, err := registerHTTP(context.Background(), "user", "pass", url, "", "")
	if err != nil {
		t.Fatalf("registerHTTP failed: %v", err)
	}

	if resp.Token != "testtoken123" {
		t.Errorf("unexpected token: got %q want %q", resp.Token, "testtoken123")
	}
}

func TestRegisterGRPC(t *testing.T) {
	cleanupRegisterTestDBFile(t)

	grpcSrv, addr := startRegisterTestGRPCServer(t)
	defer grpcSrv.GracefulStop()

	time.Sleep(100 * time.Millisecond)

	url := "grpc://" + addr

	resp, err := registerGRPC(context.Background(), "user", "pass", url, "", "")
	if err != nil {
		t.Fatalf("registerGRPC failed: %v", err)
	}

	if resp.Token != "testtoken123" {
		t.Errorf("unexpected token: got %q want %q", resp.Token, "testtoken123")
	}
}

func TestRegisterRegisterCommand(t *testing.T) {
	// Backup original funcs
	origRegisterHTTPFunc := registerHTTPFunc
	origRegisterGRPCFunc := registerGRPCFunc
	defer func() {
		registerHTTPFunc = origRegisterHTTPFunc
		registerGRPCFunc = origRegisterGRPCFunc
	}()

	// Mock HTTP and gRPC registration to return tokens
	registerHTTPFunc = func(ctx context.Context, username, password, authURL, tlsCertFile, tlsKeyFile string) (*models.AuthResponse, error) {
		return &models.AuthResponse{Token: "http-token"}, nil
	}
	registerGRPCFunc = func(ctx context.Context, username, password, authURL, tlsCertFile, tlsKeyFile string) (*models.AuthResponse, error) {
		return &models.AuthResponse{Token: "grpc-token"}, nil
	}

	root := &cobra.Command{Use: "root"}
	RegisterRegisterCommand(root)

	root.SetArgs([]string{
		"register",
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
	require.Contains(t, output, "http-token")
}
