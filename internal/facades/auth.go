package facades

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-resty/resty/v2"
	"github.com/sbilibin2017/gophkeeper/internal/models"
	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc"
)

// RegisterHTTPFacade реализует регистрацию пользователя через HTTP API.
type RegisterHTTPFacade struct {
	client *resty.Client
}

// NewRegisterHTTPFacade создает новый экземпляр RegisterHTTPFacade.
func NewRegisterHTTPFacade(client *resty.Client) *RegisterHTTPFacade {
	return &RegisterHTTPFacade{client: client}
}

// Register отправляет запрос на регистрацию пользователя через HTTP.
// При успешной регистрации возвращает токен аутентификации.
// В случае ошибки возвращает её.
func (f *RegisterHTTPFacade) Register(
	ctx context.Context,
	req *models.AuthRequest,
) (string, error) {
	var respBody models.AuthResponse

	resp, err := f.client.R().
		SetContext(ctx).
		SetBody(req).
		SetResult(&respBody).
		Post("/register")
	if err != nil {
		return "", fmt.Errorf("server unavailable: %w", err)
	}

	if resp.IsError() {
		return "", errors.New("registration error: " + resp.Status())
	}

	if respBody.Token == "" {
		return "", errors.New("token not received from server")
	}

	return respBody.Token, nil
}

// RegisterGRPCFacade реализует регистрацию пользователя через gRPC API.
type RegisterGRPCFacade struct {
	client pb.AuthServiceClient
}

// NewRegisterGRPCFacade создает новый экземпляр RegisterGRPCFacade.
func NewRegisterGRPCFacade(client pb.AuthServiceClient) *RegisterGRPCFacade {
	return &RegisterGRPCFacade{client: client}
}

// Register отправляет запрос на регистрацию пользователя через gRPC.
// При успешной регистрации возвращает токен аутентификации.
// В случае ошибки возвращает её.
func (f *RegisterGRPCFacade) Register(ctx context.Context, req *models.AuthRequest) (string, error) {
	resp, err := f.client.Register(ctx, &pb.AuthRequest{
		Username: req.Username,
		Password: req.Password,
	})
	if err != nil {
		return "", err
	}

	if resp.Token == "" {
		return "", errors.New("token not received from server")
	}

	return resp.Token, nil
}

// LoginHTTPFacade реализует аутентификацию пользователя через HTTP API.
type LoginHTTPFacade struct {
	client *resty.Client
}

// NewLoginHTTPFacade создает новый экземпляр LoginHTTPFacade.
func NewLoginHTTPFacade(client *resty.Client) *LoginHTTPFacade {
	return &LoginHTTPFacade{client: client}
}

// Login отправляет запрос на аутентификацию пользователя через HTTP.
// При успешной аутентификации возвращает токен.
// В случае ошибки возвращает её.
func (f *LoginHTTPFacade) Login(ctx context.Context, req *models.AuthRequest) (string, error) {
	var respBody models.AuthResponse

	resp, err := f.client.R().
		SetContext(ctx).
		SetBody(req).
		SetResult(&respBody).
		Post("/login")
	if err != nil {
		return "", err
	}

	if resp.IsError() {
		return "", errors.New("authentication error: " + resp.Status())
	}

	if respBody.Token == "" {
		return "", errors.New("token not received from server")
	}

	return respBody.Token, nil
}

// LoginGRPCFacade реализует аутентификацию пользователя через gRPC API.
type LoginGRPCFacade struct {
	client pb.AuthServiceClient
}

// NewLoginGRPCFacade создает новый экземпляр LoginGRPCFacade.
func NewLoginGRPCFacade(client pb.AuthServiceClient) *LoginGRPCFacade {
	return &LoginGRPCFacade{client: client}
}

// Login отправляет запрос на аутентификацию пользователя через gRPC.
// При успешной аутентификации возвращает токен.
// В случае ошибки возвращает её.
func (f *LoginGRPCFacade) Login(ctx context.Context, req *models.AuthRequest) (string, error) {
	resp, err := f.client.Login(ctx, &pb.AuthRequest{
		Username: req.Username,
		Password: req.Password,
	})
	if err != nil {
		return "", err
	}

	if resp.Token == "" {
		return "", errors.New("token not received from server")
	}

	return resp.Token, nil
}
