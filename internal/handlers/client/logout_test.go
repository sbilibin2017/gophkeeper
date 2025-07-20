package client

import (
	"context"
	"net"
	"net/http"
	"os"
	"testing"
	"time"

	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc/auth"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

// --- HTTP Test Server for Logout ---

func startLogoutTestHTTPServer(t *testing.T) (*http.Server, string) {
	handler := http.NewServeMux()
	handler.HandleFunc("/logout", func(w http.ResponseWriter, r *http.Request) {
		// Simulate successful logout response
		w.WriteHeader(http.StatusOK)
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

// --- gRPC Test Server for Logout ---

type logoutTestAuthServer struct {
	pb.UnimplementedAuthServiceServer
}

// Fix: Change request type from *pb.LogoutRequest to *emptypb.Empty
func (s *logoutTestAuthServer) Logout(ctx context.Context, req *emptypb.Empty) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil // return empty protobuf message on success
}

func startLogoutTestGRPCServer(t *testing.T) (*grpc.Server, string) {
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	grpcSrv := grpc.NewServer()
	pb.RegisterAuthServiceServer(grpcSrv, &logoutTestAuthServer{})

	go func() {
		if err := grpcSrv.Serve(lis); err != nil {
			t.Logf("gRPC server error: %v", err)
		}
	}()

	return grpcSrv, lis.Addr().String()
}

// --- Cleanup helper ---

func cleanupLogoutTestDBFile(t *testing.T) {
	t.Helper()
	t.Cleanup(func() {
		err := os.Remove("client.db")
		if err != nil && !os.IsNotExist(err) {
			t.Errorf("failed to remove client.db: %v", err)
		}
	})
}

// --- Tests ---

func TestLogoutHTTP(t *testing.T) {
	cleanupLogoutTestDBFile(t)

	srv, url := startLogoutTestHTTPServer(t)
	defer srv.Shutdown(context.Background())

	time.Sleep(50 * time.Millisecond) // wait for server to start

	err := LogoutHTTP(context.Background(), "dummy-token", url, "", "")
	require.NoError(t, err)
}

func TestLogoutGRPC(t *testing.T) {
	cleanupLogoutTestDBFile(t)

	grpcSrv, addr := startLogoutTestGRPCServer(t)
	defer grpcSrv.GracefulStop()

	time.Sleep(50 * time.Millisecond)

	url := "grpc://" + addr

	err := LogoutGRPC(context.Background(), "dummy-token", url, "", "")
	require.NoError(t, err)
}
