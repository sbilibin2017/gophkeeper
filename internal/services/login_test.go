package services

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"

	"github.com/sbilibin2017/gophkeeper/internal/models"
	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc"
)

// --- HTTPLoginService: integration test with httptest.Server ---

func TestHTTPLoginService_Login(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/login" || r.Method != http.MethodPost {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := resty.New()
	client.SetHostURL(server.URL)

	service := NewHTTPLoginService(client)

	err := service.Login(context.Background(), &models.Credentials{
		Username: "user",
		Password: "pass",
	})

	assert.NoError(t, err)
}

// --- gRPC Login service and test server ---

type testLoginServer struct {
	pb.UnimplementedLoginServiceServer
	shouldFail bool
}

func (s *testLoginServer) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	if s.shouldFail {
		return nil, errors.New("login failed")
	}
	return &pb.LoginResponse{}, nil
}

func TestGRPCLoginService_Login(t *testing.T) {
	const bufSize = 1024 * 1024
	lis := bufconn.Listen(bufSize)

	srv := grpc.NewServer()
	testSrv := &testLoginServer{}
	pb.RegisterLoginServiceServer(srv, testSrv)

	errCh := make(chan error, 1)
	go func() {
		if err := srv.Serve(lis); err != nil {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		t.Fatalf("failed to start gRPC server: %v", err)
	default:
	}

	defer srv.Stop()

	ctx := context.Background()
	conn, err := grpc.Dial(
		"bufnet",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return lis.Dial()
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	assert.NoError(t, err)
	defer conn.Close()

	client := pb.NewLoginServiceClient(conn)
	service := NewGRPCLoginService(client)

	t.Run("successful login", func(t *testing.T) {
		testSrv.shouldFail = false

		err := service.Login(ctx, &models.Credentials{
			Username: "grpcuser",
			Password: "grpcpass",
		})
		assert.NoError(t, err)
	})

	t.Run("failed login", func(t *testing.T) {
		testSrv.shouldFail = true

		err := service.Login(ctx, &models.Credentials{
			Username: "grpcuser",
			Password: "grpcpass",
		})
		assert.Error(t, err)
	})
}
