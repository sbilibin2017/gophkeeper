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

	"github.com/sbilibin2017/gophkeeper/internal/models"
	"github.com/sbilibin2017/gophkeeper/internal/services"
	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc"
	"google.golang.org/grpc/status"
)

// Тестовый HTTP сервер для RegisterHTTP
func TestRegisterHTTP(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/register" || r.Method != http.MethodPost {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"token":"test-token"}`))
	})
	server := httptest.NewServer(handler)
	defer server.Close()

	client := resty.New().SetBaseURL(server.URL)

	token, err := services.RegisterHTTP(context.Background(), client, &models.UsernamePassword{
		Username: "user",
		Password: "pass",
	})

	require.NoError(t, err)
	require.Equal(t, "test-token", token)
}

// Тестовый gRPC сервер, реализующий RegisterServiceServer
type testRegisterServer struct {
	pb.UnimplementedRegisterServiceServer
}

func (s *testRegisterServer) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	if req.Username == "user" && req.Password == "pass" {
		return &pb.RegisterResponse{Token: "grpc-test-token"}, nil
	}
	return &pb.RegisterResponse{Error: "invalid credentials"}, nil
}

func TestRegisterGRPC(t *testing.T) {
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	srv := grpc.NewServer()
	pb.RegisterRegisterServiceServer(srv, &testRegisterServer{})

	go srv.Serve(lis)
	defer srv.Stop()

	conn, err := grpc.Dial(lis.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	defer conn.Close()

	client := pb.NewRegisterServiceClient(conn)

	token, err := services.RegisterGRPC(context.Background(), client, &models.UsernamePassword{
		Username: "user",
		Password: "pass",
	})

	require.NoError(t, err)
	require.Equal(t, "grpc-test-token", token)
}

// Тесты ошибок HTTP клиента
func TestRegisterHTTP_Errors(t *testing.T) {
	// Сервер с ошибкой 500
	server500 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server500.Close()

	client := resty.New().SetBaseURL(server500.URL)

	_, err := services.RegisterHTTP(context.Background(), client, &models.UsernamePassword{Username: "user", Password: "pass"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "server returned an error response")

	// Сервер без токена
	serverNoToken := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"token":""}`))
	}))
	defer serverNoToken.Close()

	client = resty.New().SetBaseURL(serverNoToken.URL)

	_, err = services.RegisterHTTP(context.Background(), client, &models.UsernamePassword{Username: "user", Password: "pass"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "token not received in server response")

	// Некорректный адрес сервера (симуляция ошибки запроса)
	client = resty.New().SetBaseURL("http://invalid-host")

	_, err = services.RegisterHTTP(context.Background(), client, &models.UsernamePassword{Username: "user", Password: "pass"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to perform HTTP request")
}

// Тесты ошибок gRPC клиента
type errorServer struct {
	pb.UnimplementedRegisterServiceServer
}

func (s *errorServer) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	switch req.Username {
	case "error":
		return nil, status.Errorf(13, "some grpc error") // код 13 = internal
	case "errorresp":
		return &pb.RegisterResponse{Error: "invalid user"}, nil
	case "emptytoken":
		return &pb.RegisterResponse{Token: ""}, nil
	default:
		return &pb.RegisterResponse{Token: "token123"}, nil
	}
}

func TestRegisterGRPC_Errors(t *testing.T) {
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	srv := grpc.NewServer()
	pb.RegisterRegisterServiceServer(srv, &errorServer{})
	go srv.Serve(lis)
	defer srv.Stop()

	conn, err := grpc.NewClient(lis.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	defer conn.Close()

	client := pb.NewRegisterServiceClient(conn)

	_, err = services.RegisterGRPC(context.Background(), client, &models.UsernamePassword{Username: "error", Password: "pass"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to perform gRPC request")

	_, err = services.RegisterGRPC(context.Background(), client, &models.UsernamePassword{Username: "errorresp", Password: "pass"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "registration error: invalid user")

	_, err = services.RegisterGRPC(context.Background(), client, &models.UsernamePassword{Username: "emptytoken", Password: "pass"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "token not received from gRPC server")
}
