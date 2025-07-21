package auth

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"os"
	"testing"
	"time"

	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc/auth"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// minimalGRPCAuthServer starts a gRPC test server with a simple AuthService implementation
func minimalGRPCAuthServer(t *testing.T, certFile, keyFile string) (*grpc.Server, string) {
	// Load TLS cert
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		t.Fatalf("failed to load key pair: %v", err)
	}
	creds := credentials.NewServerTLSFromCert(&cert)

	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}

	server := grpc.NewServer(grpc.Creds(creds))

	// Implement the AuthService server
	pb.RegisterAuthServiceServer(server, &testAuthService{})

	go func() {
		if err := server.Serve(lis); err != nil {
			t.Logf("gRPC server error: %v", err)
		}
	}()

	return server, lis.Addr().String()
}

type testAuthService struct {
	pb.UnimplementedAuthServiceServer
}

func (s *testAuthService) Login(ctx context.Context, req *pb.AuthRequest) (*pb.AuthResponse, error) {
	// Return a single token for consistency with HTTP test
	return &pb.AuthResponse{
		Token: "test_token",
	}, nil
}

func TestNewRunGRPC_Integration(t *testing.T) {
	// Remove old DB if exists
	if err := os.Remove("client.db"); err != nil && !os.IsNotExist(err) {
		t.Fatalf("failed to remove client.db: %v", err)
	}
	// Schedule deletion after test
	defer func() {
		if err := os.Remove("client.db"); err != nil && !os.IsNotExist(err) {
			t.Fatalf("failed to remove client.db: %v", err)
		}
	}()

	certFile, keyFile := generateSelfSignedCert(t)
	defer os.Remove(certFile)
	defer os.Remove(keyFile)

	server, addr := minimalGRPCAuthServer(t, certFile, keyFile)
	defer server.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	authURL := fmt.Sprintf("grpc://%s", addr)
	run := NewRunGRPC(authURL, certFile, keyFile)

	resp, err := run(ctx, "dummyuser", "dummypass")

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "test_token", resp.Token)
}
