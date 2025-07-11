package facades

import (
	"context"
	"fmt"

	"github.com/go-resty/resty/v2"
	"github.com/sbilibin2017/gophkeeper/internal/models"
	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc"
	"google.golang.org/grpc"
)

// RegisterHTTPFacade - фасад для регистрации через HTTP с resty
type RegisterHTTPFacade struct {
	client *resty.Client
}

// NewRegisterHTTPFacade создаёт новый HTTP фасад для регистрации
func NewRegisterHTTPFacade(client *resty.Client) (*RegisterHTTPFacade, error) {
	return &RegisterHTTPFacade{
		client: client,
	}, nil
}

// Register отправляет запрос на регистрацию и возвращает токен или ошибку
func (r *RegisterHTTPFacade) Register(ctx context.Context, secret *models.UsernamePassword) (string, error) {
	var result struct {
		Token string `json:"token"`
	}

	resp, err := r.client.R().
		SetContext(ctx).
		SetBody(secret).
		SetResult(&result).
		Post("/register")

	if err != nil {
		return "", fmt.Errorf("failed to perform HTTP request: %w", err)
	}

	if resp.IsError() {
		return "", fmt.Errorf("server returned an error response: %s", resp.Status())
	}

	if result.Token == "" {
		return "", fmt.Errorf("token not received in server response")
	}

	return result.Token, nil
}

// RegisterGRPCFacade - фасад для регистрации через gRPC
type RegisterGRPCFacade struct {
	client pb.RegisterServiceClient
}

// NewRegisterGRPCFacade создаёт новый gRPC фасад для регистрации из *grpc.ClientConn
func NewRegisterGRPCFacade(conn *grpc.ClientConn) (*RegisterGRPCFacade, error) {
	return &RegisterGRPCFacade{
		client: pb.NewRegisterServiceClient(conn),
	}, nil
}

// Register отправляет gRPC-запрос на регистрацию и возвращает токен или ошибку
func (r *RegisterGRPCFacade) Register(ctx context.Context, secret *models.UsernamePassword) (string, error) {
	req := &pb.RegisterRequest{
		Username: secret.Username,
		Password: secret.Password,
	}

	resp, err := r.client.Register(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to perform gRPC request: %w", err)
	}

	if resp.Error != "" {
		return "", fmt.Errorf("registration error: %s", resp.Error)
	}

	if resp.Token == "" {
		return "", fmt.Errorf("token not received from gRPC server")
	}

	return resp.Token, nil
}
