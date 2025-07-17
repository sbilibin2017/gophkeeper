package client

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/sbilibin2017/gophkeeper/internal/models"
	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

// --- HTTP server for tests ---

func authRunTestHTTPServer(t *testing.T) (*http.Server, string) {
	mux := http.NewServeMux()
	mux.HandleFunc("/auth", func(w http.ResponseWriter, r *http.Request) {
		var req models.AuthRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		resp := models.AuthResponse{Token: "test-token-for-" + req.Username}
		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(resp)
		require.NoError(t, err)
	})

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	srv := &http.Server{Handler: mux}

	go func() {
		err := srv.Serve(listener)
		if err != nil && err != http.ErrServerClosed {
			t.Errorf("HTTP server error: %v", err)
		}
	}()

	return srv, listener.Addr().String()
}

// --- gRPC server for tests ---

const authBufSize = 1024 * 1024

type authTestAuthServer struct {
	pb.UnimplementedAuthServiceServer
}

func (s *authTestAuthServer) Auth(ctx context.Context, req *pb.AuthRequest) (*pb.AuthResponse, error) {
	return &pb.AuthResponse{Token: "test-token-for-" + req.Username}, nil
}

func authRunTestGRPCServer(t *testing.T) (*grpc.Server, *bufconn.Listener) {
	listener := bufconn.Listen(authBufSize)
	grpcServer := grpc.NewServer()
	pb.RegisterAuthServiceServer(grpcServer, &authTestAuthServer{})

	go func() {
		if err := grpcServer.Serve(listener); err != nil {
			t.Errorf("gRPC server error: %v", err)
		}
	}()

	return grpcServer, listener
}

// Dialer for bufconn gRPC client
func authBufDialer(listener *bufconn.Listener) func(context.Context, string) (net.Conn, error) {
	return func(ctx context.Context, _ string) (net.Conn, error) {
		return listener.Dial()
	}
}

// --- Tests ---

func TestAuthHTTP(t *testing.T) {
	srv, addr := authRunTestHTTPServer(t)
	defer srv.Close()

	client := resty.New()
	client.SetBaseURL("http://" + addr)

	ctx := context.Background()
	req := &models.AuthRequest{Username: "user1", Password: "pass"}

	resp, err := AuthHTTP(ctx, client, req)
	require.NoError(t, err)
	require.Equal(t, "test-token-for-user1", resp.Token)
}

func TestAuthGRPC(t *testing.T) {
	grpcServer, listener := authRunTestGRPCServer(t)
	defer grpcServer.Stop()

	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet",
		grpc.WithContextDialer(authBufDialer(listener)), // bufconn dialer
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)
	defer conn.Close()

	grpcClient := pb.NewAuthServiceClient(conn)
	req := &models.AuthRequest{Username: "user1", Password: "pass"}

	resp, err := AuthGRPC(ctx, grpcClient, req)
	require.NoError(t, err)
	require.Equal(t, "test-token-for-user1", resp.Token)
}
