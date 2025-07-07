package services

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-resty/resty/v2"

	"github.com/sbilibin2017/gophkeeper/internal/models"
	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc"
)

// Registerer описывает интерфейс сервиса регистрации пользователя.
type Registerer interface {
	// Register выполняет регистрацию пользователя с указанными учётными данными.
	Register(ctx context.Context, creds *models.Credentials) error
}

// RegisterContextService предоставляет обёртку для вызова регистрации через Registerer.
type RegisterContextService struct {
	registerer Registerer
}

// NewRegisterContextService создаёт новый экземпляр RegisterContextService.
func NewRegisterContextService() *RegisterContextService {
	return &RegisterContextService{}
}

// SetContext задаёт реализацию интерфейса Registerer для последующих вызовов.
func (r *RegisterContextService) SetContext(registerer Registerer) {
	r.registerer = registerer
}

// Register вызывает регистрацию пользователя через установленный Registerer.
// Возвращает ошибку, если Registerer не установлен или регистрация не удалась.
func (r *RegisterContextService) Register(
	ctx context.Context,
	creds *models.Credentials,
) error {
	if r.registerer == nil {
		return fmt.Errorf("registerer not set")
	}
	return r.registerer.Register(ctx, creds)
}

// HTTPRegisterService реализует регистрацию через HTTP API.
type HTTPRegisterService struct {
	client *resty.Client
}

// NewHTTPRegisterService создаёт новый HTTPRegisterService с заданным HTTP клиентом.
func NewHTTPRegisterService(client *resty.Client) *HTTPRegisterService {
	return &HTTPRegisterService{client: client}
}

// Register отправляет HTTP POST запрос для регистрации пользователя.
// Возвращает ошибку, если запрос не удался или сервер вернул статус отличный от 200 OK.
func (r *HTTPRegisterService) Register(
	ctx context.Context,
	creds *models.Credentials,
) error {
	resp, err := r.client.R().
		SetContext(ctx).
		SetHeader("Content-Type", "application/json").
		SetBody(creds).
		Post("/register")

	if err != nil {
		return err
	}

	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("registration failed: %s", resp.String())
	}

	return nil
}

// GRPCRegisterService реализует регистрацию через gRPC сервис.
type GRPCRegisterService struct {
	client pb.RegisterServiceClient
}

// NewGRPCRegisterService создаёт новый GRPCRegisterService с заданным gRPC клиентом.
func NewGRPCRegisterService(client pb.RegisterServiceClient) *GRPCRegisterService {
	return &GRPCRegisterService{client: client}
}

// Register отправляет gRPC запрос для регистрации пользователя.
// Возвращает ошибку, если запрос не удался.
func (r *GRPCRegisterService) Register(ctx context.Context, creds *models.Credentials) error {
	_, err := r.client.Register(ctx, &pb.RegisterRequest{
		Username: creds.Username,
		Password: creds.Password,
	})
	if err != nil {
		return err
	}

	return nil
}
