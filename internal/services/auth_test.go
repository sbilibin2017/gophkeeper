package services

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"

	"github.com/sbilibin2017/gophkeeper/internal/models"

	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc"
)

// --- HTTP Server Handlers (fake) ---

func TestLoginHTTP(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		// Возвращаем тестовый токен в json
		w.Header().Set("Content-Type", "application/json")
		resp := struct {
			Token string `json:"token"`
		}{
			Token: "test-jwt-token",
		}
		json.NewEncoder(w).Encode(resp)
	})

	server := &http.Server{Addr: ":8081", Handler: mux}
	go server.ListenAndServe()
	defer server.Close()

	// Даем серверу немного времени запуститься
	time.Sleep(100 * time.Millisecond)

	client := resty.New().SetBaseURL("http://localhost:8081")

	token, err := LoginHTTP(context.Background(), client, &models.Credentials{
		Username: "testuser",
		Password: "testpass",
	})

	require.NoError(t, err)
	assert.Equal(t, "test-jwt-token", token)
}

func TestRegisterHTTP(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		// Возвращаем тестовый токен в json
		w.Header().Set("Content-Type", "application/json")
		resp := struct {
			Token string `json:"token"`
		}{
			Token: "register-jwt-token",
		}
		json.NewEncoder(w).Encode(resp)
	})

	server := &http.Server{Addr: ":8082", Handler: mux}
	go server.ListenAndServe()
	defer server.Close()

	time.Sleep(100 * time.Millisecond)

	client := resty.New().SetBaseURL("http://localhost:8082")

	token, err := RegisterHTTP(context.Background(), client, &models.Credentials{
		Username: "newuser",
		Password: "newpass",
	})

	require.NoError(t, err)
	assert.Equal(t, "register-jwt-token", token)
}

// --- gRPC Server Implementation (fake) ---

type testLoginServiceServer struct {
	pb.UnimplementedLoginServiceServer
}

func (s *testLoginServiceServer) Login(ctx context.Context, req *pb.Credentials) (*pb.LoginResponse, error) {
	if req.Username == "" || req.Password == "" {
		return &pb.LoginResponse{Error: "missing credentials"}, nil
	}
	return &pb.LoginResponse{Token: "grpc-login-token"}, nil
}

type testRegisterServiceServer struct {
	pb.UnimplementedRegisterServiceServer
}

func (s *testRegisterServiceServer) Register(ctx context.Context, req *pb.Credentials) (*pb.RegisterResponse, error) {
	if req.Username == "" || req.Password == "" {
		return &pb.RegisterResponse{Error: "missing credentials"}, nil
	}
	return &pb.RegisterResponse{Token: "grpc-register-token"}, nil
}

func startTestGRPCServer(t *testing.T) (addr string, stop func()) {
	lis, err := net.Listen("tcp", "127.0.0.1:0") // любой свободный порт
	require.NoError(t, err)

	grpcServer := grpc.NewServer()
	pb.RegisterLoginServiceServer(grpcServer, &testLoginServiceServer{})
	pb.RegisterRegisterServiceServer(grpcServer, &testRegisterServiceServer{})
	reflection.Register(grpcServer)

	go grpcServer.Serve(lis)

	return lis.Addr().String(), func() {
		grpcServer.Stop()
		lis.Close()
	}
}

func TestLoginGRPC(t *testing.T) {
	addr, stop := startTestGRPCServer(t)
	defer stop()

	conn, err := grpc.Dial(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)
	defer conn.Close()

	client := pb.NewLoginServiceClient(conn)

	token, err := LoginGRPC(context.Background(), client, &models.Credentials{
		Username: "grpcuser",
		Password: "grpcpass",
	})

	require.NoError(t, err)
	assert.Equal(t, "grpc-login-token", token)
}

func TestRegisterGRPC(t *testing.T) {
	addr, stop := startTestGRPCServer(t)
	defer stop()

	conn, err := grpc.Dial(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)
	defer conn.Close()

	client := pb.NewRegisterServiceClient(conn)

	token, err := RegisterGRPC(context.Background(), client, &models.Credentials{
		Username: "grpcnewuser",
		Password: "grpcnewpass",
	})

	require.NoError(t, err)
	assert.Equal(t, "grpc-register-token", token)
}
