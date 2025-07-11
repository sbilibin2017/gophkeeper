package facades

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
	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc"
)

// Тестовый HTTP сервер для RegisterHTTPFacade
func TestRegisterHTTPFacade_Register(t *testing.T) {
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
	facade, err := NewRegisterHTTPFacade(client)
	require.NoError(t, err)

	token, err := facade.Register(context.Background(), &models.UsernamePassword{
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

func TestRegisterGRPCFacade_Register(t *testing.T) {
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	srv := grpc.NewServer()
	pb.RegisterRegisterServiceServer(srv, &testRegisterServer{})

	go srv.Serve(lis)
	defer srv.Stop()

	conn, err := grpc.NewClient(lis.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	defer conn.Close()

	facade, err := NewRegisterGRPCFacade(conn)
	require.NoError(t, err)

	token, err := facade.Register(context.Background(), &models.UsernamePassword{
		Username: "user",
		Password: "pass",
	})

	require.NoError(t, err)
	require.Equal(t, "grpc-test-token", token)
}

// Тест ошибки HTTP клиента: ошибка запроса, статус ошибки, пустой токен
func TestRegisterHTTPFacade_Register_Errors(t *testing.T) {
	// Сервер, возвращающий 500 ошибку
	server500 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server500.Close()

	client := resty.New().SetBaseURL(server500.URL)
	facade, err := NewRegisterHTTPFacade(client)
	require.NoError(t, err)

	_, err = facade.Register(context.Background(), &models.UsernamePassword{Username: "user", Password: "pass"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "server returned an error response")

	// Сервер возвращает 200, но без токена
	serverNoToken := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"token":""}`))
	}))
	defer serverNoToken.Close()

	client = resty.New().SetBaseURL(serverNoToken.URL)
	facade, err = NewRegisterHTTPFacade(client)
	require.NoError(t, err)

	_, err = facade.Register(context.Background(), &models.UsernamePassword{Username: "user", Password: "pass"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "token not received in server response")

	// Клиент с невалидным адресом (симулируем ошибку запроса)
	client = resty.New().SetBaseURL("http://invalid-host")
	facade, err = NewRegisterHTTPFacade(client)
	require.NoError(t, err)

	_, err = facade.Register(context.Background(), &models.UsernamePassword{Username: "user", Password: "pass"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to perform HTTP request")
}

type errorServer struct {
	pb.UnimplementedRegisterServiceServer
}

func (s *errorServer) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	if req.Username == "error" {
		return nil, grpc.Errorf(13, "some grpc error") // code 13 = internal
	}
	if req.Username == "errorresp" {
		return &pb.RegisterResponse{Error: "invalid user"}, nil
	}
	if req.Username == "emptytoken" {
		return &pb.RegisterResponse{Token: ""}, nil
	}
	return &pb.RegisterResponse{Token: "token123"}, nil
}

func TestRegisterGRPCFacade_Register_Errors(t *testing.T) {
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	srv := grpc.NewServer()
	pb.RegisterRegisterServiceServer(srv, &errorServer{})
	go srv.Serve(lis)
	defer srv.Stop()

	conn, err := grpc.NewClient(lis.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	defer conn.Close()

	facade, err := NewRegisterGRPCFacade(conn)
	require.NoError(t, err)

	// Ошибка запроса gRPC
	_, err = facade.Register(context.Background(), &models.UsernamePassword{Username: "error", Password: "pass"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to perform gRPC request")

	// Ошибка в ответе gRPC
	_, err = facade.Register(context.Background(), &models.UsernamePassword{Username: "errorresp", Password: "pass"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "registration error: invalid user")

	// Пустой токен в ответе
	_, err = facade.Register(context.Background(), &models.UsernamePassword{Username: "emptytoken", Password: "pass"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "token not received from gRPC server")
}
