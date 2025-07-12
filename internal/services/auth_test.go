package services_test

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"

	"github.com/sbilibin2017/gophkeeper/internal/models"
	"github.com/sbilibin2017/gophkeeper/internal/services"
	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc"
)

// Тестовый HTTP сервер для AuthHTTP
func TestAuthHTTP(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/auth" || r.Method != http.MethodPost {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"token":"test-token"}`))
	})
	server := httptest.NewServer(handler)
	defer server.Close()

	client := resty.New().SetBaseURL(server.URL)

	token, err := services.AuthHTTP(context.Background(), client, &models.UsernamePassword{
		Username: "user",
		Password: "pass",
	})

	require.NoError(t, err)
	require.Equal(t, "test-token", token)
}

// Тестовый gRPC сервер, реализующий AuthServiceServer
type testAuthServer struct {
	pb.UnimplementedAuthServiceServer
}

func (s *testAuthServer) Auth(ctx context.Context, req *pb.AuthRequest) (*pb.AuthResponse, error) {
	if req.Username == "user" && req.Password == "pass" {
		return &pb.AuthResponse{Token: "grpc-test-token"}, nil
	}
	return &pb.AuthResponse{Error: "invalid credentials"}, nil
}

func TestAuthGRPC(t *testing.T) {
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	srv := grpc.NewServer()
	pb.RegisterAuthServiceServer(srv, &testAuthServer{})

	go srv.Serve(lis)
	defer srv.Stop()

	conn, err := grpc.Dial(lis.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	defer conn.Close()

	client := pb.NewAuthServiceClient(conn)

	token, err := services.AuthGRPC(context.Background(), client, &models.UsernamePassword{
		Username: "user",
		Password: "pass",
	})

	require.NoError(t, err)
	require.Equal(t, "grpc-test-token", token)
}

// Тесты ошибок HTTP клиента для AuthHTTP
func TestAuthHTTP_Errors(t *testing.T) {
	// Сервер с ошибкой 500
	server500 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server500.Close()

	client := resty.New().SetBaseURL(server500.URL)

	_, err := services.AuthHTTP(context.Background(), client, &models.UsernamePassword{Username: "user", Password: "pass"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "server returned an error response")

	// Сервер без токена
	serverNoToken := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"token":""}`))
	}))
	defer serverNoToken.Close()

	client = resty.New().SetBaseURL(serverNoToken.URL)

	_, err = services.AuthHTTP(context.Background(), client, &models.UsernamePassword{Username: "user", Password: "pass"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "token not received in server response")

	// Некорректный адрес сервера (симуляция ошибки запроса)
	client = resty.New().SetBaseURL("http://invalid-host")

	_, err = services.AuthHTTP(context.Background(), client, &models.UsernamePassword{Username: "user", Password: "pass"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to perform HTTP request")
}

// Тесты ошибок gRPC клиента для AuthGRPC
type errorAuthServer struct {
	pb.UnimplementedAuthServiceServer
}

func (s *errorAuthServer) Auth(ctx context.Context, req *pb.AuthRequest) (*pb.AuthResponse, error) {
	switch req.Username {
	case "error":
		return nil, status.Errorf(13, "some grpc error") // код 13 = internal
	case "errorresp":
		return &pb.AuthResponse{Error: "invalid user"}, nil
	case "emptytoken":
		return &pb.AuthResponse{Token: ""}, nil
	default:
		return &pb.AuthResponse{Token: "token123"}, nil
	}
}

func TestAuthGRPC_Errors(t *testing.T) {
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	srv := grpc.NewServer()
	pb.RegisterAuthServiceServer(srv, &errorAuthServer{})
	go srv.Serve(lis)
	defer srv.Stop()

	conn, err := grpc.Dial(lis.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	defer conn.Close()

	client := pb.NewAuthServiceClient(conn)

	_, err = services.AuthGRPC(context.Background(), client, &models.UsernamePassword{Username: "error", Password: "pass"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to perform gRPC request")

	_, err = services.AuthGRPC(context.Background(), client, &models.UsernamePassword{Username: "errorresp", Password: "pass"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "authentication error: invalid user")

	_, err = services.AuthGRPC(context.Background(), client, &models.UsernamePassword{Username: "emptytoken", Password: "pass"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "token not received from gRPC server")
}
