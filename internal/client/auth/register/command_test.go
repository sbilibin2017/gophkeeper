package register

import (
	"bytes"
	"context"
	"net"
	"net/http"
	"testing"
	"time"

	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc" // adjust import path as needed
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

// Dummy gRPC AuthService implementation for testing
type testAuthServiceServer struct {
	pb.UnimplementedAuthServiceServer
}

func (s *testAuthServiceServer) Register(ctx context.Context, req *pb.AuthRequest) (*pb.AuthResponse, error) {
	return &pb.AuthResponse{Token: "dummy-grpc-token"}, nil
}

// Start a simple HTTP test server for registration
func startTestHTTPServer(t *testing.T) (*http.Server, net.Listener) {
	handler := http.NewServeMux()
	handler.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Authorization", "Bearer real-http-token")
		w.WriteHeader(http.StatusOK)
	})

	server := &http.Server{
		Handler: handler,
	}

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	assert.NoError(t, err)

	go func() {
		err := server.Serve(ln)
		// http.ErrServerClosed is expected on shutdown, ignore it
		if err != nil && err != http.ErrServerClosed {
			assert.NoError(t, err)
		}
	}()

	time.Sleep(50 * time.Millisecond) // wait for server to start

	return server, ln
}

// Start a simple gRPC test server for registration
func startTestGRPCServer2(t *testing.T) (*grpc.Server, net.Listener) {
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	assert.NoError(t, err)

	grpcServer := grpc.NewServer()
	pb.RegisterAuthServiceServer(grpcServer, &testAuthServiceServer{})

	go func() {
		err := grpcServer.Serve(lis)
		assert.NoError(t, err)
	}()

	time.Sleep(50 * time.Millisecond) // wait for server to start

	return grpcServer, lis
}

func TestNewCommand_HTTP_Integration(t *testing.T) {
	server, ln := startTestHTTPServer(t)
	defer func() {
		// Graceful shutdown to avoid "http: Server closed" error
		_ = server.Shutdown(context.Background())
		_ = ln.Close()
	}()

	cmd := NewCommand()
	cmd.SetArgs([]string{
		"--server-url", "http://" + ln.Addr().String(),
		"--username", "testuser",
		"--password", "testpass",
	})

	outBuf := &bytes.Buffer{}
	cmd.SetOut(outBuf)

	err := cmd.Execute()
	assert.NoError(t, err)
	assert.Contains(t, outBuf.String(), "real-http-token")
}

func TestNewCommand_GRPC_Integration(t *testing.T) {
	grpcServer, lis := startTestGRPCServer2(t)
	defer grpcServer.Stop()

	cmd := NewCommand()
	cmd.SetArgs([]string{
		"--server-url", "grpc://" + lis.Addr().String(),
		"--username", "testuser",
		"--password", "testpass",
	})

	outBuf := &bytes.Buffer{}
	cmd.SetOut(outBuf)

	err := cmd.Execute()
	assert.NoError(t, err)
	assert.Contains(t, outBuf.String(), "dummy-grpc-token")
}

func TestNewCommand_UnsupportedScheme(t *testing.T) {
	cmd := NewCommand()
	cmd.SetArgs([]string{
		"--server-url", "ftp://localhost",
		"--username", "testuser",
		"--password", "testpass",
	})

	err := cmd.Execute()
	assert.Error(t, err)

}
