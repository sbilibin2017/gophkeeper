package facades

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"

	"github.com/sbilibin2017/gophkeeper/internal/models"
	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc"
)

const (
	httpBaseURL = "http://localhost:8080" // твой HTTP сервер
	grpcAddress = "localhost:50051"       // твой gRPC сервер
	timeout     = 5 * time.Second
)

func TestRegisterHTTPFacade_Register(t *testing.T) {
	// Мок HTTP сервер
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/register" {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		var req models.AuthRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(models.AuthResponse{
			Token: "test-token",
		})
	}))
	defer ts.Close()

	// Создаем resty клиент с базовым URL тестового сервера
	restyClient := resty.New().
		SetBaseURL(ts.URL).
		SetTimeout(5 * time.Second)

	client := NewRegisterHTTPFacade(restyClient)

	token, err := client.Register(context.Background(), &models.AuthRequest{
		Username: "user1",
		Password: "pass1",
	})

	assert.NoError(t, err)
	assert.Equal(t, "test-token", token)
}

const bufSize = 1024 * 1024

var lis *bufconn.Listener

// Реализация мок-сервера AuthService
type mockAuthServer struct {
	pb.UnimplementedAuthServiceServer
}

func (s *mockAuthServer) Register(ctx context.Context, req *pb.AuthRequest) (*pb.AuthResponse, error) {
	// Можно добавить проверку req, если нужно
	return &pb.AuthResponse{Token: "grpc-test-token"}, nil
}

func (s *mockAuthServer) Login(ctx context.Context, req *pb.AuthRequest) (*pb.AuthResponse, error) {
	return &pb.AuthResponse{Token: "grpc-test-token"}, nil
}

func dialer() func(context.Context, string) (net.Conn, error) {
	return func(ctx context.Context, s string) (net.Conn, error) {
		return lis.Dial()
	}
}

func TestRegisterGRPCFacade_Register(t *testing.T) {
	lis = bufconn.Listen(bufSize)
	s := grpc.NewServer()
	pb.RegisterAuthServiceServer(s, &mockAuthServer{})

	errCh := make(chan error, 1)
	go func() {
		if err := s.Serve(lis); err != nil {
			errCh <- err
		}
	}()

	defer s.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, "bufnet",
		grpc.WithContextDialer(dialer()),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	assert.NoError(t, err)
	defer conn.Close()

	client := pb.NewAuthServiceClient(conn)
	facade := NewRegisterGRPCFacade(client)

	req := &models.AuthRequest{
		Username: "testuser",
		Password: "testpass",
	}

	token, err := facade.Register(ctx, req)
	assert.NoError(t, err)
	assert.Equal(t, "grpc-test-token", token)

	// Проверяем, не было ли ошибок в серверной горутине
	select {
	case err := <-errCh:
		t.Fatalf("gRPC server error: %v", err)
	default:
		// Нет ошибок — всё ок
	}
}

func TestLoginHTTPFacade_Login(t *testing.T) {
	// Мок HTTP сервер для /login
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/login" {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		var req models.AuthRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(models.AuthResponse{
			Token: "login-test-token",
		})
	}))
	defer ts.Close()

	restyClient := resty.New().
		SetBaseURL(ts.URL).
		SetTimeout(5 * time.Second)

	client := NewLoginHTTPFacade(restyClient)

	token, err := client.Login(context.Background(), &models.AuthRequest{
		Username: "user2",
		Password: "pass2",
	})

	assert.NoError(t, err)
	assert.Equal(t, "login-test-token", token)
}

func TestLoginGRPCFacade_Login(t *testing.T) {
	lis = bufconn.Listen(bufSize)
	s := grpc.NewServer()
	pb.RegisterAuthServiceServer(s, &mockAuthServer{})

	errCh := make(chan error, 1)
	go func() {
		if err := s.Serve(lis); err != nil {
			errCh <- err
		}
	}()

	defer s.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, "bufnet",
		grpc.WithContextDialer(dialer()),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	assert.NoError(t, err)
	defer conn.Close()

	client := pb.NewAuthServiceClient(conn)
	facade := NewLoginGRPCFacade(client)

	req := &models.AuthRequest{
		Username: "testuser2",
		Password: "testpass2",
	}

	token, err := facade.Login(ctx, req)
	assert.NoError(t, err)
	assert.Equal(t, "grpc-test-token", token)

	select {
	case err := <-errCh:
		t.Fatalf("gRPC server error: %v", err)
	default:
	}
}
