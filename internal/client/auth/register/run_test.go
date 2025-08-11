package register

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"

	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc" // adjust import to your proto package
)

// --- HTTP Tests ---

func TestRunHTTP_Success(t *testing.T) {
	// Setup test HTTP server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && r.URL.Path == "/register" {
			// Simulate successful registration returning auth token in header
			w.Header().Set("Authorization", "Bearer test-token-123")
			w.WriteHeader(http.StatusOK)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	token, err := RunHTTP(context.Background(), ts.URL, "testuser", "testpass")
	assert.NoError(t, err)
	assert.Equal(t, "test-token-123", token)
}

func TestRunHTTP_Failure(t *testing.T) {
	// No HTTP server, connection should fail or unauthorized
	_, err := RunHTTP(context.Background(), "http://localhost:8080", "invaliduser", "invalidpass")
	assert.Error(t, err)
}

// --- gRPC Mock Server and Tests ---

// mockAuthServer implements the gRPC Auth service interface for testing.
type mockAuthServer struct {
	pb.UnimplementedAuthServiceServer
}

func (s *mockAuthServer) Register(ctx context.Context, req *pb.AuthRequest) (*pb.AuthResponse, error) {
	return &pb.AuthResponse{Token: "test-token"}, nil
}

// startTestGRPCServer starts a test gRPC server and returns its address and a cleanup function.
func startTestGRPCServer(t *testing.T) (string, func()) {
	lis, err := net.Listen("tcp", "localhost:0") // random free port
	assert.NoError(t, err)

	server := grpc.NewServer()
	pb.RegisterAuthServiceServer(server, &mockAuthServer{})

	go func() {
		_ = server.Serve(lis)
	}()

	return lis.Addr().String(), func() {
		server.Stop()
		lis.Close()
	}
}

func TestRunGRPC_Success(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	addr, cleanup := startTestGRPCServer(t)
	defer cleanup()

	token, err := RunGRPC(ctx, addr, "testuser", "testpass")
	assert.NoError(t, err)
	assert.Equal(t, "test-token", token)
}

func TestRunGRPC_Failure(t *testing.T) {
	_, err := RunGRPC(context.Background(), "localhost:65535", "invaliduser", "invalidpass")
	assert.Error(t, err)
}

func TestRunGRPC_BadAddress(t *testing.T) {
	_, err := RunGRPC(context.Background(), "invalid:0000", "anyuser", "anypass")
	assert.Error(t, err)
}
