package services

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"

	"github.com/sbilibin2017/gophkeeper/internal/models"
	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc"
)

func TestRegisterContextService_Register(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRegisterer := NewMockRegisterer(ctrl)
	service := NewRegisterContextService()
	service.SetContext(mockRegisterer)

	ctx := context.Background()
	creds := &models.Credentials{
		Username: "testuser",
		Password: "testpass",
	}

	t.Run("successful registration", func(t *testing.T) {
		mockRegisterer.EXPECT().
			Register(ctx, creds).
			Return(nil)

		err := service.Register(ctx, creds)
		assert.NoError(t, err)
	})

	t.Run("registration error", func(t *testing.T) {
		expectedErr := errors.New("register error")
		mockRegisterer.EXPECT().
			Register(ctx, creds).
			Return(expectedErr)

		err := service.Register(ctx, creds)
		assert.EqualError(t, err, expectedErr.Error())
	})

	t.Run("registerer not set", func(t *testing.T) {
		serviceWithoutRegisterer := NewRegisterContextService()
		err := serviceWithoutRegisterer.Register(ctx, creds)
		assert.EqualError(t, err, "registerer not set")
	})
}

// --- HTTPRegisterService: интеграционный тест с httptest.Server ---

func TestHTTPRegisterService_Register(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/register" || r.Method != http.MethodPost {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := resty.New()
	client.SetHostURL(server.URL)

	service := NewHTTPRegisterService(client)

	err := service.Register(context.Background(), &models.Credentials{
		Username: "user",
		Password: "pass",
	})

	assert.NoError(t, err)
}

// --- gRPC сервис и сервер для теста ---

type testRegisterServer struct {
	pb.UnimplementedRegisterServiceServer
	shouldFail bool
}

func (s *testRegisterServer) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	if s.shouldFail {
		return nil, errors.New("registration failed")
	}
	return &pb.RegisterResponse{Error: ""}, nil
}

func TestGRPCRegisterService_Register(t *testing.T) {
	const bufSize = 1024 * 1024
	lis := bufconn.Listen(bufSize)

	srv := grpc.NewServer()
	testSrv := &testRegisterServer{}
	pb.RegisterRegisterServiceServer(srv, testSrv)

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

	client := pb.NewRegisterServiceClient(conn)
	service := NewGRPCRegisterService(client)

	t.Run("successful registration", func(t *testing.T) {
		testSrv.shouldFail = false

		err := service.Register(ctx, &models.Credentials{
			Username: "grpcuser",
			Password: "grpcpass",
		})
		assert.NoError(t, err)
	})

	t.Run("failed registration", func(t *testing.T) {
		testSrv.shouldFail = true

		err := service.Register(ctx, &models.Credentials{
			Username: "grpcuser",
			Password: "grpcpass",
		})
		assert.Error(t, err)
	})
}
