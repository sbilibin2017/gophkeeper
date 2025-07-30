package facades

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"

	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc"
)

// --- HTTP Tests ---

func TestAuthHTTPFacade_RegisterAndLogin(t *testing.T) {
	handler := http.NewServeMux()

	handler.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		w.Header().Set("Authorization", "Bearer register-token-for-"+req.Username)
		w.WriteHeader(http.StatusOK)
	})

	handler.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		w.Header().Set("Authorization", "Bearer login-token-for-"+req.Username)
		w.WriteHeader(http.StatusOK)
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	client := NewAuthHTTPFacade(newRestyClientWithBaseURL(server.URL))
	ctx := context.Background()

	// Test Register
	registerResp, err := client.Register(ctx, "user1", "pass")
	require.NoError(t, err)
	require.NotNil(t, registerResp)
	assert.Equal(t, "register-token-for-user1", *registerResp)

	// Test Login
	loginResp, err := client.Login(ctx, "user1", "pass")
	require.NoError(t, err)
	require.NotNil(t, loginResp)
	assert.Equal(t, "login-token-for-user1", *loginResp)
}

// helper function for resty client with base URL for tests
func newRestyClientWithBaseURL(baseURL string) *resty.Client {
	client := resty.New()
	client.SetBaseURL(baseURL)
	return client
}

// --- gRPC Tests ---

// mockAuthServiceServer implements pb.AuthServiceServer for testing
type mockAuthServiceServer struct {
	pb.UnimplementedAuthServiceServer
}

func (m *mockAuthServiceServer) Register(ctx context.Context, req *pb.AuthRequest) (*pb.AuthResponse, error) {
	return &pb.AuthResponse{Token: "register-token-for-" + req.Username}, nil
}

func (m *mockAuthServiceServer) Login(ctx context.Context, req *pb.AuthRequest) (*pb.AuthResponse, error) {
	return &pb.AuthResponse{Token: "login-token-for-" + req.Username}, nil
}

func TestAuthGRPCFacade_RegisterAndLogin(t *testing.T) {
	lis, err := net.Listen("tcp", "127.0.0.1:0") // choose random available port
	require.NoError(t, err)

	grpcServer := grpc.NewServer()
	pb.RegisterAuthServiceServer(grpcServer, &mockAuthServiceServer{})

	go grpcServer.Serve(lis)
	defer grpcServer.Stop()

	conn, err := grpc.Dial(lis.Addr().String(), grpc.WithInsecure())
	require.NoError(t, err)
	defer conn.Close()

	client := NewAuthGRPCFacade(conn)
	ctx := context.Background()

	// Test Register
	registerResp, err := client.Register(ctx, "user1", "pass")
	require.NoError(t, err)
	require.NotNil(t, registerResp)
	assert.Equal(t, "register-token-for-user1", *registerResp)

	// Test Login
	loginResp, err := client.Login(ctx, "user1", "pass")
	require.NoError(t, err)
	require.NotNil(t, loginResp)
	assert.Equal(t, "login-token-for-user1", *loginResp)
}

func TestAuthHTTPFacade_ErrorCases(t *testing.T) {
	// HTTP 500 error simulation
	handler := http.NewServeMux()
	handler.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "internal error", http.StatusInternalServerError)
	})
	handler.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	client := NewAuthHTTPFacade(newRestyClientWithBaseURL(server.URL))
	ctx := context.Background()

	// Register HTTP error
	_, err := client.Register(ctx, "user1", "pass")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "register request returned error")

	// Login HTTP error
	_, err = client.Login(ctx, "user1", "pass")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "login request returned error")

	// Network error (invalid URL)
	badClient := NewAuthHTTPFacade(newRestyClientWithBaseURL("http://invalid.localhost"))
	_, err = badClient.Register(ctx, "user1", "pass")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "register request failed")

	_, err = badClient.Login(ctx, "user1", "pass")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "login request failed")
}
