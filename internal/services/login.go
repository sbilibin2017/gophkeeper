package services

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-resty/resty/v2"

	"github.com/sbilibin2017/gophkeeper/internal/models"
	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc"
)

// Loginer описывает интерфейс сервиса аутентификации пользователя.
type Loginer interface {
	// Login выполняет аутентификацию пользователя с указанными учётными данными.
	Login(ctx context.Context, creds *models.Credentials) error
}

// LoginContextService предоставляет обёртку для вызова аутентификации через Loginer.
type LoginContextService struct {
	loginer Loginer
}

// NewLoginContextService создаёт новый экземпляр LoginContextService.
func NewLoginContextService() *LoginContextService {
	return &LoginContextService{}
}

// SetContext задаёт реализацию интерфейса Loginer для последующих вызовов.
func (l *LoginContextService) SetContext(loginer Loginer) {
	l.loginer = loginer
}

// Login вызывает аутентификацию пользователя через установленный Loginer.
// Возвращает ошибку, если Loginer не установлен или аутентификация не удалась.
func (l *LoginContextService) Login(ctx context.Context, creds *models.Credentials) error {
	if l.loginer == nil {
		return fmt.Errorf("loginer not set")
	}
	return l.loginer.Login(ctx, creds)
}

// HTTPLoginService реализует аутентификацию через HTTP API.
type HTTPLoginService struct {
	client *resty.Client
}

// NewHTTPLoginService создаёт новый HTTPLoginService с заданным HTTP клиентом.
func NewHTTPLoginService(client *resty.Client) *HTTPLoginService {
	return &HTTPLoginService{client: client}
}

// Login отправляет HTTP POST запрос для аутентификации пользователя.
// Возвращает ошибку, если запрос не удался или сервер вернул статус отличный от 200 OK.
func (l *HTTPLoginService) Login(ctx context.Context, creds *models.Credentials) error {
	resp, err := l.client.R().
		SetContext(ctx).
		SetHeader("Content-Type", "application/json").
		SetBody(creds).
		Post("/login")

	if err != nil {
		return err
	}

	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("login failed: %s", resp.String())
	}

	return nil
}

// GRPCLoginService реализует аутентификацию через gRPC сервис.
type GRPCLoginService struct {
	client pb.LoginServiceClient
}

// NewGRPCLoginService создаёт новый GRPCLoginService с заданным gRPC клиентом.
func NewGRPCLoginService(client pb.LoginServiceClient) *GRPCLoginService {
	return &GRPCLoginService{client: client}
}

// Login отправляет gRPC запрос для аутентификации пользователя.
// Возвращает ошибку, если запрос не удался.
func (l *GRPCLoginService) Login(ctx context.Context, creds *models.Credentials) error {
	_, err := l.client.Login(ctx, &pb.LoginRequest{
		Username: creds.Username,
		Password: creds.Password,
	})
	if err != nil {
		return err
	}

	return nil
}
