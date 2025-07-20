package handlers

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/sbilibin2017/gophkeeper/internal/models"
	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc/auth"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

// --- HTTP Test Server for Login ---

func startLoginTestHTTPServer(t *testing.T) (*http.Server, string) {
	handler := http.NewServeMux()
	handler.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		resp := models.AuthResponse{Token: "testtoken123"}
		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode(resp)
		require.NoError(t, err)
	})

	srv := &http.Server{Handler: handler}

	ln, err := net.Listen("tcp", "127.0.0.1:0") // random port
	require.NoError(t, err)

	go func() {
		if err := srv.Serve(ln); err != nil && err != http.ErrServerClosed {
			t.Logf("HTTP server error: %v", err)
		}
	}()

	return srv, "http://" + ln.Addr().String()
}

// --- gRPC Test Server for Login ---

type loginTestAuthServer struct {
	pb.UnimplementedAuthServiceServer
}

func (s *loginTestAuthServer) Login(ctx context.Context, req *pb.AuthRequest) (*pb.AuthResponse, error) {
	return &pb.AuthResponse{Token: "testtoken123"}, nil
}

func startLoginTestGRPCServer(t *testing.T) (*grpc.Server, string) {
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	grpcSrv := grpc.NewServer()
	pb.RegisterAuthServiceServer(grpcSrv, &loginTestAuthServer{})

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

	// Wait for server startup
	time.Sleep(50 * time.Millisecond)

	resp, err := LoginHTTP(context.Background(), "user", "pass", url, "", "")
	require.NoError(t, err)
	require.Equal(t, "testtoken123", resp.Token)
}

func TestLoginGRPC(t *testing.T) {
	cleanupLoginTestDBFile(t)

	grpcSrv, addr := startLoginTestGRPCServer(t)
	defer grpcSrv.GracefulStop()

	time.Sleep(50 * time.Millisecond)

	url := "grpc://" + addr

	resp, err := LoginGRPC(context.Background(), "user", "pass", url, "", "")
	require.NoError(t, err)
	require.Equal(t, "testtoken123", resp.Token)
}
