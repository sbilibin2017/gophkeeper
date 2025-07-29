package facades

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/sbilibin2017/gophkeeper/inernal/models"
	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

const bufSize = 1024 * 1024

// Mock gRPC server setup
func setupGRPCServer(t *testing.T) (*grpc.ClientConn, func()) {
	lis := bufconn.Listen(bufSize)
	s := grpc.NewServer()

	pb.RegisterAuthServiceServer(s, &mockAuthServer{})

	go func() {
		err := s.Serve(lis)
		require.NoError(t, err)
	}()

	conn, err := grpc.DialContext(context.Background(), "bufnet",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return lis.Dial()
		}),
		grpc.WithInsecure(),
	)
	require.NoError(t, err)

	return conn, func() {
		conn.Close()
		s.Stop()
	}
}

// gRPC mock implementation
type mockAuthServer struct {
	pb.UnimplementedAuthServiceServer
}

func (m *mockAuthServer) Register(ctx context.Context, req *pb.AuthRegisterRequest) (*pb.AuthResponse, error) {
	return &pb.AuthResponse{Token: "grpc-register-token"}, nil
}

func (m *mockAuthServer) Login(ctx context.Context, req *pb.AuthLoginRequest) (*pb.AuthResponse, error) {
	return &pb.AuthResponse{Token: "grpc-login-token"}, nil
}

// ------------------- TESTS -------------------

func TestAuthHTTP_Register(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/register/", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"token":"mock-token"}`))
	}))
	defer srv.Close()

	client := resty.New().SetBaseURL(srv.URL)
	auth := NewAuthHTTP(client)

	req := &models.AuthRegisterRequest{
		Username: "user1",
		Password: "pass1",
	}

	resp, err := auth.Register(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, "mock-token", resp.Token)
}

func TestAuthHTTP_Login(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/login/", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"token":"login-token"}`))
	}))
	defer srv.Close()

	client := resty.New().SetBaseURL(srv.URL)
	auth := NewAuthHTTP(client)

	req := &models.AuthLoginRequest{
		Username: "user1",
		Password: "pass1",
	}

	resp, err := auth.Login(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, "login-token", resp.Token)
}

func TestAuthGRPC_Register(t *testing.T) {
	conn, cleanup := setupGRPCServer(t)
	defer cleanup()

	auth := NewAuthGRPC(conn)

	req := &models.AuthRegisterRequest{
		Username: "user1",
		Password: "pass1",
	}

	resp, err := auth.Register(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, "grpc-register-token", resp.Token)
}

func TestAuthGRPC_Login(t *testing.T) {
	conn, cleanup := setupGRPCServer(t)
	defer cleanup()

	auth := NewAuthGRPC(conn)

	req := &models.AuthLoginRequest{
		Username: "user1",
		Password: "pass1",
	}

	resp, err := auth.Login(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, "grpc-login-token", resp.Token)
}
