package handlers

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/sbilibin2017/gophkeeper/internal/client/models"
	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc/auth"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

// --- HTTP Test Server ---

func startRegisterTestHTTPServer(t *testing.T) (*http.Server, string) {
	handler := http.NewServeMux()
	handler.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		resp := models.AuthResponse{Token: "testtoken123"}
		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode(resp)
		require.NoError(t, err)
	})

	srv := &http.Server{Handler: handler}

	ln, err := net.Listen("tcp", "127.0.0.1:0") // listen on random port
	require.NoError(t, err)

	go func() {
		if err := srv.Serve(ln); err != nil && err != http.ErrServerClosed {
			t.Logf("HTTP server error: %v", err)
		}
	}()

	return srv, "http://" + ln.Addr().String()
}

// --- gRPC Test Server ---

type registerTestAuthServer struct {
	pb.UnimplementedAuthServiceServer
}

func (s *registerTestAuthServer) Register(ctx context.Context, req *pb.AuthRequest) (*pb.AuthResponse, error) {
	return &pb.AuthResponse{Token: "testtoken123"}, nil
}

func startRegisterTestGRPCServer(t *testing.T) (*grpc.Server, string) {
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	grpcSrv := grpc.NewServer()
	pb.RegisterAuthServiceServer(grpcSrv, &registerTestAuthServer{})

	go func() {
		if err := grpcSrv.Serve(lis); err != nil {
			t.Logf("gRPC server error: %v", err)
		}
	}()

	return grpcSrv, lis.Addr().String()
}

// --- Cleanup helper ---

func cleanupRegisterTestDBFile(t *testing.T) {
	t.Helper()
	t.Cleanup(func() {
		err := os.Remove("client.db")
		if err != nil && !os.IsNotExist(err) {
			t.Errorf("failed to remove client.db: %v", err)
		}
	})
}

// --- Tests ---

func TestRegisterHTTP(t *testing.T) {
	cleanupRegisterTestDBFile(t)

	srv, url := startRegisterTestHTTPServer(t)
	defer srv.Shutdown(context.Background())

	// Allow server time to start
	time.Sleep(50 * time.Millisecond)

	resp, err := RegisterHTTP(context.Background(), "user", "pass", url, "", "")
	require.NoError(t, err)
	require.Equal(t, "testtoken123", resp.Token)
}

func TestRegisterGRPC(t *testing.T) {
	cleanupRegisterTestDBFile(t)

	grpcSrv, addr := startRegisterTestGRPCServer(t)
	defer grpcSrv.GracefulStop()

	time.Sleep(50 * time.Millisecond)

	url := "grpc://" + addr

	resp, err := RegisterGRPC(context.Background(), "user", "pass", url, "", "")
	require.NoError(t, err)
	require.Equal(t, "testtoken123", resp.Token)
}
