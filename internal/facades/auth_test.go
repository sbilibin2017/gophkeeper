package facades

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/sbilibin2017/gophkeeper/internal/models"
	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

const bufSize = 1024 * 1024

// --- HTTP Facade Integration Tests ---

func TestAuthHTTPFacade_Register_Login(t *testing.T) {
	// HTTP test server responds with Authorization header on /register and /login
	mux := http.NewServeMux()
	mux.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Authorization", "Bearer register-token")
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Authorization", "Bearer login-token")
		w.WriteHeader(http.StatusOK)
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	client := resty.New().SetBaseURL(server.URL)
	facade := NewAuthHTTPFacade(client)

	ctx := context.Background()
	req := &models.AuthRequest{Username: "user", Password: "pass"}

	// Register
	resp, err := facade.Register(ctx, req)
	assert.NoError(t, err)
	assert.Equal(t, "register-token", resp.Token)

	// Login
	resp, err = facade.Login(ctx, req)
	assert.NoError(t, err)
	assert.Equal(t, "login-token", resp.Token)
}

// HTTP error tests

func TestAuthHTTPFacade_Register_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}))
	defer server.Close()

	client := resty.New().SetBaseURL(server.URL)
	facade := NewAuthHTTPFacade(client)

	_, err := facade.Register(context.Background(), &models.AuthRequest{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "register request returned error")
}

func TestAuthHTTPFacade_Login_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer server.Close()

	client := resty.New().SetBaseURL(server.URL)
	facade := NewAuthHTTPFacade(client)

	_, err := facade.Login(context.Background(), &models.AuthRequest{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "login request returned error")
}

func TestAuthHTTPFacade_Register_MissingAuthHeader(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		// No Authorization header set
	}))
	defer server.Close()

	client := resty.New().SetBaseURL(server.URL)
	facade := NewAuthHTTPFacade(client)

	_, err := facade.Register(context.Background(), &models.AuthRequest{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "authorization header missing")
}

func TestAuthHTTPFacade_Login_InvalidAuthHeaderFormat(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Authorization", "InvalidFormat token123")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := resty.New().SetBaseURL(server.URL)
	facade := NewAuthHTTPFacade(client)

	_, err := facade.Login(context.Background(), &models.AuthRequest{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "authorization header format invalid")
}

func TestAuthHTTPFacade_Register_NetworkError(t *testing.T) {
	client := resty.New().SetBaseURL("http://invalid.invalid")
	facade := NewAuthHTTPFacade(client)

	_, err := facade.Register(context.Background(), &models.AuthRequest{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "register request failed")
}

func TestAuthHTTPFacade_Login_NetworkError(t *testing.T) {
	client := resty.New().SetBaseURL("http://invalid.invalid")
	facade := NewAuthHTTPFacade(client)

	_, err := facade.Login(context.Background(), &models.AuthRequest{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "login request failed")
}

// --- gRPC Facade Integration Tests ---

func startGRPCTestServer(t *testing.T) (*grpc.ClientConn, func()) {
	lis := bufconn.Listen(bufSize)

	srv := grpc.NewServer()
	pb.RegisterAuthServiceServer(srv, &mockAuthServiceServer{})

	go func() {
		if err := srv.Serve(lis); err != nil {
			t.Fatalf("gRPC server exited with error: %v", err)
		}
	}()

	conn, err := grpc.DialContext(context.Background(), "bufnet",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return lis.Dial()
		}),
		grpc.WithInsecure(),
	)
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}

	return conn, func() {
		conn.Close()
		srv.Stop()
	}
}

type mockAuthServiceServer struct {
	pb.UnimplementedAuthServiceServer
}

func (m *mockAuthServiceServer) Register(ctx context.Context, req *pb.AuthRequest) (*pb.AuthResponse, error) {
	return &pb.AuthResponse{Token: "grpc-register-token"}, nil
}

func (m *mockAuthServiceServer) Login(ctx context.Context, req *pb.AuthRequest) (*pb.AuthResponse, error) {
	return &pb.AuthResponse{Token: "grpc-login-token"}, nil
}

func TestAuthGRPCFacade_Register_Login(t *testing.T) {
	conn, cleanup := startGRPCTestServer(t)
	defer cleanup()

	facade := NewAuthGRPCFacade(conn)
	ctx := context.Background()
	req := &models.AuthRequest{Username: "user", Password: "pass"}

	// Register
	resp, err := facade.Register(ctx, req)
	assert.NoError(t, err)
	assert.Equal(t, "grpc-register-token", resp.Token)

	// Login
	resp, err = facade.Login(ctx, req)
	assert.NoError(t, err)
	assert.Equal(t, "grpc-login-token", resp.Token)
}

// gRPC error tests

func startGRPCErrorServer(t *testing.T, errToReturn error) (*grpc.ClientConn, func()) {
	lis := bufconn.Listen(bufSize)
	srv := grpc.NewServer()

	pb.RegisterAuthServiceServer(srv, &errorAuthServiceServer{err: errToReturn})

	go func() {
		if err := srv.Serve(lis); err != nil {
			t.Fatalf("gRPC server error: %v", err)
		}
	}()

	conn, err := grpc.DialContext(context.Background(), "bufnet",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return lis.Dial()
		}),
		grpc.WithInsecure(),
	)
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}

	return conn, func() {
		conn.Close()
		srv.Stop()
	}
}

type errorAuthServiceServer struct {
	pb.UnimplementedAuthServiceServer
	err error
}

func (e *errorAuthServiceServer) Register(ctx context.Context, req *pb.AuthRequest) (*pb.AuthResponse, error) {
	return nil, e.err
}

func (e *errorAuthServiceServer) Login(ctx context.Context, req *pb.AuthRequest) (*pb.AuthResponse, error) {
	return nil, e.err
}

func TestAuthGRPCFacade_Register_Error(t *testing.T) {
	conn, cleanup := startGRPCErrorServer(t, status.Error(codes.Internal, "internal error"))
	defer cleanup()

	facade := NewAuthGRPCFacade(conn)
	_, err := facade.Register(context.Background(), &models.AuthRequest{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "internal error")
}

func TestAuthGRPCFacade_Login_Error(t *testing.T) {
	conn, cleanup := startGRPCErrorServer(t, errors.New("network failure"))
	defer cleanup()

	facade := NewAuthGRPCFacade(conn)
	_, err := facade.Login(context.Background(), &models.AuthRequest{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "network failure")
}

func TestAuthHTTPFacade_Register_MissingAuthorizationHeader(t *testing.T) {
	// Start a test HTTP server that returns 200 OK but NO Authorization header
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Intentionally omit Authorization header
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := resty.New().SetBaseURL(server.URL)
	facade := NewAuthHTTPFacade(client)

	_, err := facade.Register(context.Background(), &models.AuthRequest{Username: "test", Password: "test"})
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "authorization header missing")
	}
}

func TestAuthHTTPFacade_Register_InvalidAuthHeaderFormat(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set Authorization header but in wrong format (missing "Bearer " prefix)
		w.Header().Set("Authorization", "InvalidTokenFormat123")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := resty.New().SetBaseURL(server.URL)
	facade := NewAuthHTTPFacade(client)

	_, err := facade.Register(context.Background(), &models.AuthRequest{Username: "test", Password: "test"})
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "authorization header format invalid")
	}
}

func TestAuthHTTPFacade_Login_MissingAuthHeader(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		// No Authorization header
	}))
	defer server.Close()

	client := resty.New().SetBaseURL(server.URL)
	facade := NewAuthHTTPFacade(client)

	_, err := facade.Login(context.Background(), &models.AuthRequest{Username: "test", Password: "test"})
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "authorization header missing")
	}
}
