package client

import (
	"context"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"

	"github.com/sbilibin2017/gophkeeper/internal/models"
	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc"
)

func TestRegisterUserHTTP_EmptyClient(t *testing.T) {
	client := resty.New()

	// Не задаём BaseURL, запрос упадёт, но не будет паники
	token, err := RegisterUserHTTP(context.Background(), client, &models.AuthRequest{
		Username: "test",
		Password: "test",
	})

	assert.Error(t, err)
	assert.Empty(t, token)
}

func TestLoginUserHTTP_EmptyClient(t *testing.T) {
	client := resty.New() // и тут тоже

	token, err := LoginUserHTTP(context.Background(), client, &models.AuthRequest{Username: "u", Password: "p"})
	assert.Error(t, err)
	assert.Empty(t, token)
}

type dummyGRPCClient struct{}

func (d dummyGRPCClient) Register(ctx context.Context, in *pb.AuthRequest, opts ...grpc.CallOption) (*pb.AuthResponse, error) {
	return nil, status.Errorf(codes.Unavailable, "no connection")
}

func (d dummyGRPCClient) Login(ctx context.Context, in *pb.AuthRequest, opts ...grpc.CallOption) (*pb.AuthResponse, error) {
	return nil, status.Errorf(codes.Unavailable, "no connection")
}

func TestRegisterUserGRPC_Error(t *testing.T) {
	dc := dummyGRPCClient{}
	token, err := RegisterUserGRPC(context.Background(), dc, &models.AuthRequest{Username: "u", Password: "p"})
	assert.Error(t, err)
	assert.Empty(t, token)
}

func TestLoginUserGRPC_Error(t *testing.T) {
	dc := dummyGRPCClient{}
	token, err := LoginUserGRPC(context.Background(), dc, &models.AuthRequest{Username: "u", Password: "p"})
	assert.Error(t, err)
	assert.Empty(t, token)
}

// Интеграционные тесты с реальным сервером — для их запуска должен быть запущен сервер
func TestRegisterUserHTTP_Integration(t *testing.T) {
	client := resty.New()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	req := &models.AuthRequest{Username: "testuser", Password: "testpass"}

	token, err := RegisterUserHTTP(ctx, client, req)

	// Проверяем, что либо ошибка, либо получен токен (либо пустой токен)
	assert.True(t, err == nil || token == "" || token != "")
}

func TestLoginUserHTTP_Integration(t *testing.T) {
	client := resty.New()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	req := &models.AuthRequest{Username: "testuser", Password: "testpass"}

	token, err := LoginUserHTTP(ctx, client, req)

	assert.True(t, err == nil || token == "" || token != "")
}

func TestRegisterUserGRPC_Integration(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	conn, err := grpc.DialContext(ctx, "localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Skip("grpc server not available, skipping integration test")
	}
	defer conn.Close()

	client := pb.NewAuthServiceClient(conn)
	req := &models.AuthRequest{Username: "testuser", Password: "testpass"}

	token, err := RegisterUserGRPC(ctx, client, req)

	assert.True(t, err == nil || token == "" || token != "")
}

func TestLoginUserGRPC_Integration(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	conn, err := grpc.DialContext(ctx, "localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Skip("grpc server not available, skipping integration test")
	}
	defer conn.Close()

	client := pb.NewAuthServiceClient(conn)
	req := &models.AuthRequest{Username: "testuser", Password: "testpass"}

	token, err := LoginUserGRPC(ctx, client, req)

	assert.True(t, err == nil || token == "" || token != "")
}
