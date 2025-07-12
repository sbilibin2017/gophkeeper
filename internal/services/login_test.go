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

// ==== HTTP ====

func TestLoginHTTP(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/login" || r.Method != http.MethodPost {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"token":"login-token"}`))
	})
	server := httptest.NewServer(handler)
	defer server.Close()

	client := resty.New().SetBaseURL(server.URL)

	token, err := services.LoginHTTP(context.Background(), client, &models.UsernamePassword{
		Username: "user",
		Password: "pass",
	})

	require.NoError(t, err)
	require.Equal(t, "login-token", token)
}

func TestLoginHTTP_Errors(t *testing.T) {
	// Сервер с ошибкой 401
	server401 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server401.Close()

	client := resty.New().SetBaseURL(server401.URL)

	_, err := services.LoginHTTP(context.Background(), client, &models.UsernamePassword{Username: "user", Password: "pass"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "server returned an error response")

	// Сервер без токена
	serverNoToken := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"token":""}`))
	}))
	defer serverNoToken.Close()

	client = resty.New().SetBaseURL(serverNoToken.URL)

	_, err = services.LoginHTTP(context.Background(), client, &models.UsernamePassword{Username: "user", Password: "pass"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "token not received in server response")

	// Ошибочный URL
	client = resty.New().SetBaseURL("http://invalid-host")

	_, err = services.LoginHTTP(context.Background(), client, &models.UsernamePassword{Username: "user", Password: "pass"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to perform HTTP request")
}

// ==== gRPC ====

type testLoginServer struct {
	pb.UnimplementedLoginServiceServer
}

func (s *testLoginServer) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	if req.Username == "user" && req.Password == "pass" {
		return &pb.LoginResponse{Token: "grpc-login-token"}, nil
	}
	return &pb.LoginResponse{Error: "invalid credentials"}, nil
}

func TestLoginGRPC(t *testing.T) {
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	srv := grpc.NewServer()
	pb.RegisterLoginServiceServer(srv, &testLoginServer{})
	go srv.Serve(lis)
	defer srv.Stop()

	conn, err := grpc.Dial(lis.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	defer conn.Close()

	client := pb.NewLoginServiceClient(conn)

	token, err := services.LoginGRPC(context.Background(), client, &models.UsernamePassword{
		Username: "user",
		Password: "pass",
	})
	require.NoError(t, err)
	require.Equal(t, "grpc-login-token", token)
}

type errorLoginServer struct {
	pb.UnimplementedLoginServiceServer
}

func (s *errorLoginServer) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	switch req.Username {
	case "error":
		return nil, status.Errorf(13, "some grpc error") // gRPC code 13 — internal
	case "errorresp":
		return &pb.LoginResponse{Error: "login failed"}, nil
	case "emptytoken":
		return &pb.LoginResponse{Token: ""}, nil
	default:
		return &pb.LoginResponse{Token: "valid-token"}, nil
	}
}

func TestLoginGRPC_Errors(t *testing.T) {
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	srv := grpc.NewServer()
	pb.RegisterLoginServiceServer(srv, &errorLoginServer{})
	go srv.Serve(lis)
	defer srv.Stop()

	conn, err := grpc.Dial(lis.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	defer conn.Close()

	client := pb.NewLoginServiceClient(conn)

	_, err = services.LoginGRPC(context.Background(), client, &models.UsernamePassword{Username: "error", Password: "pass"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to perform gRPC request")

	_, err = services.LoginGRPC(context.Background(), client, &models.UsernamePassword{Username: "errorresp", Password: "pass"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "login error: login failed")

	_, err = services.LoginGRPC(context.Background(), client, &models.UsernamePassword{Username: "emptytoken", Password: "pass"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "token not received from gRPC server")
}
