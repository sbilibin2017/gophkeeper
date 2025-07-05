package services

import (
	"context"
	"errors"

	"github.com/sbilibin2017/gophkeeper/internal/models"
)

// Registerer определяет интерфейс для сервисов регистрации пользователя.
type Registerer interface {
	// Register выполняет регистрацию пользователя с указанными учетными данными.
	Register(ctx context.Context, creds *models.Credentials) error
}

// RegisterService содержит Registerer и делегирует вызовы регистрации ему.
type RegisterService struct {
	context Registerer
}

// NewRegisterService создаёт новый экземпляр RegisterService.
func NewRegisterService() *RegisterService {
	return &RegisterService{}
}

// SetContext устанавливает конкретную реализацию Registerer для сервиса.
func (svc *RegisterService) SetContext(r Registerer) {
	svc.context = r
}

// Register вызывает регистрацию через установленный Registerer.
// Если контекст регистрации не установлен, возвращает ошибку.
func (svc *RegisterService) Register(ctx context.Context, creds *models.Credentials) error {
	if svc.context == nil {
		return errors.New("no context set")
	}
	return svc.context.Register(ctx, creds)
}

// RegisterHTTPServiceOption — функциональный параметр для настройки RegisterHTTPService.
type RegisterHTTPServiceOption func(*RegisterHTTPService)

// RegisterHTTPService обрабатывает регистрацию пользователя по HTTP.
type RegisterHTTPService struct {
	serverURL     string
	publicKeyPath string
	hmacKey       string
}

// NewRegisterHTTPService создаёт новый RegisterHTTPService с опциями.
func NewRegisterHTTPService(opts ...RegisterHTTPServiceOption) *RegisterHTTPService {
	svc := &RegisterHTTPService{}
	for _, opt := range opts {
		opt(svc)
	}
	return svc
}

// WithHTTPServerURL задаёт URL сервера для RegisterHTTPService.
func WithHTTPServerURL(url string) RegisterHTTPServiceOption {
	return func(s *RegisterHTTPService) {
		s.serverURL = url
	}
}

// WithHTTPPublicKeyPath задаёт путь к публичному ключу для RegisterHTTPService.
func WithHTTPPublicKeyPath(path string) RegisterHTTPServiceOption {
	return func(s *RegisterHTTPService) {
		s.publicKeyPath = path
	}
}

// WithHTTPHMACKey задаёт HMAC ключ для RegisterHTTPService.
func WithHTTPHMACKey(key string) RegisterHTTPServiceOption {
	return func(s *RegisterHTTPService) {
		s.hmacKey = key
	}
}

// Register выполняет регистрацию пользователя по HTTP.
func (svc *RegisterHTTPService) Register(ctx context.Context, creds *models.Credentials) error {
	return nil
}

// RegisterGRPCServiceOption — функциональный параметр для настройки RegisterGRPCService.
type RegisterGRPCServiceOption func(*RegisterGRPCService)

// RegisterGRPCService обрабатывает регистрацию пользователя по gRPC.
type RegisterGRPCService struct {
	serverURL     string
	publicKeyPath string
	hmacKey       string
}

// NewRegisterGRPCService создаёт новый RegisterGRPCService с опциями.
func NewRegisterGRPCService(opts ...RegisterGRPCServiceOption) *RegisterGRPCService {
	svc := &RegisterGRPCService{}
	for _, opt := range opts {
		opt(svc)
	}
	return svc
}

// WithGRPCServerURL задаёт URL сервера для RegisterGRPCService.
func WithGRPCServerURL(url string) RegisterGRPCServiceOption {
	return func(s *RegisterGRPCService) {
		s.serverURL = url
	}
}

// WithGRPCPublicKeyPath задаёт путь к публичному ключу для RegisterGRPCService.
func WithGRPCPublicKeyPath(path string) RegisterGRPCServiceOption {
	return func(s *RegisterGRPCService) {
		s.publicKeyPath = path
	}
}

// WithGRPCHMACKey задаёт HMAC ключ для RegisterGRPCService.
func WithGRPCHMACKey(key string) RegisterGRPCServiceOption {
	return func(s *RegisterGRPCService) {
		s.hmacKey = key
	}
}

// Register выполняет регистрацию пользователя по gRPC.
func (svc *RegisterGRPCService) Register(ctx context.Context, creds *models.Credentials) error {
	return nil
}
