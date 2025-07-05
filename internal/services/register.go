package services

import (
	"context"
	"errors"

	"github.com/sbilibin2017/gophkeeper/internal/models"
)

// Registerer defines the interface for user registration services.
type Registerer interface {
	// Register performs user registration with the provided credentials.
	Register(ctx context.Context, creds *models.Credentials) error
}

// RegisterService holds a Registerer and delegates registration calls to it.
type RegisterService struct {
	context Registerer
}

// NewRegisterService creates a new instance of RegisterService.
func NewRegisterService() *RegisterService {
	return &RegisterService{}
}

// SetContext sets a specific Registerer implementation for the service.
func (svc *RegisterService) SetContext(r Registerer) {
	svc.context = r
}

// Register calls registration via the set Registerer.
// Returns an error if the registration context is not set.
func (svc *RegisterService) Register(ctx context.Context, creds *models.Credentials) error {
	if svc.context == nil {
		return errors.New("no context set")
	}
	return svc.context.Register(ctx, creds)
}

// RegisterHTTPServiceOption — functional option for configuring RegisterHTTPService.
type RegisterHTTPServiceOption func(*RegisterHTTPService)

// RegisterHTTPService handles user registration over HTTP.
type RegisterHTTPService struct {
	serverURL     string
	publicKeyPath string
	hmacKey       string
}

// NewRegisterHTTPService creates a new RegisterHTTPService with options.
func NewRegisterHTTPService(opts ...RegisterHTTPServiceOption) *RegisterHTTPService {
	svc := &RegisterHTTPService{}
	for _, opt := range opts {
		opt(svc)
	}
	return svc
}

// WithHTTPServerURL sets the server URL for RegisterHTTPService.
func WithHTTPServerURL(url string) RegisterHTTPServiceOption {
	return func(s *RegisterHTTPService) {
		s.serverURL = url
	}
}

// WithHTTPPublicKeyPath sets the public key path for RegisterHTTPService.
func WithHTTPPublicKeyPath(path string) RegisterHTTPServiceOption {
	return func(s *RegisterHTTPService) {
		s.publicKeyPath = path
	}
}

// WithHTTPHMACKey sets the HMAC key for RegisterHTTPService.
func WithHTTPHMACKey(key string) RegisterHTTPServiceOption {
	return func(s *RegisterHTTPService) {
		s.hmacKey = key
	}
}

// Register performs user registration over HTTP.
func (svc *RegisterHTTPService) Register(ctx context.Context, creds *models.Credentials) error {
	return nil
}

// RegisterGRPCServiceOption — functional option for configuring RegisterGRPCService.
type RegisterGRPCServiceOption func(*RegisterGRPCService)

// RegisterGRPCService handles user registration over gRPC.
type RegisterGRPCService struct {
	serverURL     string
	publicKeyPath string
	hmacKey       string
}

// NewRegisterGRPCService creates a new RegisterGRPCService with options.
func NewRegisterGRPCService(opts ...RegisterGRPCServiceOption) *RegisterGRPCService {
	svc := &RegisterGRPCService{}
	for _, opt := range opts {
		opt(svc)
	}
	return svc
}

// WithGRPCServerURL sets the server URL for RegisterGRPCService.
func WithGRPCServerURL(url string) RegisterGRPCServiceOption {
	return func(s *RegisterGRPCService) {
		s.serverURL = url
	}
}

// WithGRPCPublicKeyPath sets the public key path for RegisterGRPCService.
func WithGRPCPublicKeyPath(path string) RegisterGRPCServiceOption {
	return func(s *RegisterGRPCService) {
		s.publicKeyPath = path
	}
}

// WithGRPCHMACKey sets the HMAC key for RegisterGRPCService.
func WithGRPCHMACKey(key string) RegisterGRPCServiceOption {
	return func(s *RegisterGRPCService) {
		s.hmacKey = key
	}
}

// Register performs user registration over gRPC.
func (svc *RegisterGRPCService) Register(ctx context.Context, creds *models.Credentials) error {
	return nil
}
