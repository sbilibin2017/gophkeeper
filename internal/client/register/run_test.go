package register

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/sbilibin2017/gophkeeper/internal/models"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"

	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc" // adjust import to your proto package
)

// --- HTTP Tests ---

func TestRunHTTP_Success(t *testing.T) {
	// Create a test HTTP server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check method and path
		if r.Method == http.MethodPost && r.URL.Path == "/register" {
			// Respond with an Authorization header to simulate successful registration
			w.Header().Set("Authorization", "Bearer test-token-123")
			w.WriteHeader(http.StatusOK)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	req := &models.AuthRequest{
		Username: "testuser",
		Password: "testpass",
	}

	resp, err := RunHTTP(context.Background(), ts.URL, req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "test-token-123", resp.Token)
}

func TestRunHTTP_Failure(t *testing.T) {
	ctx := context.Background()

	req := &models.AuthRequest{
		Username: "invaliduser",
		Password: "invalidpass",
	}

	// Assuming no server running here to cause failure:
	serverURL := "http://localhost:8080"

	resp, err := RunHTTP(ctx, serverURL, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
}

// --- gRPC Mock Server and Tests ---

// mockAuthServer implements your gRPC Auth service interface for testing.
type mockAuthServer struct {
	pb.UnimplementedAuthServiceServer // embed for forward compatibility
}

func (s *mockAuthServer) Register(ctx context.Context, req *pb.AuthRequest) (*pb.AuthResponse, error) {
	// Return a fixed token for any request
	return &pb.AuthResponse{Token: "test-token"}, nil
}

// startTestGRPCServer starts a test gRPC server and returns its address and a cleanup function.
func startTestGRPCServer(t *testing.T) (string, func()) {
	lis, err := net.Listen("tcp", "localhost:0") // use random free port
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

	req := &models.AuthRequest{
		Username: "testuser",
		Password: "testpass",
	}

	// Pass address WITHOUT grpc:// prefix
	resp, err := RunGRPC(ctx, addr, req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "test-token", resp.Token)
}

func TestRunGRPC_Failure(t *testing.T) {
	ctx := context.Background()

	req := &models.AuthRequest{
		Username: "invaliduser",
		Password: "invalidpass",
	}

	// Connect to random unused port to cause failure
	serverURL := "localhost:65535"

	resp, err := RunGRPC(ctx, serverURL, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestRunGRPC_BadAddress(t *testing.T) {
	ctx := context.Background()

	req := &models.AuthRequest{
		Username: "anyuser",
		Password: "anypass",
	}

	// Invalid address format
	serverURL := "invalid:0000"

	resp, err := RunGRPC(ctx, serverURL, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
}
