package grpc

import (
	"context"

	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc"
)

// Registerer defines the interface for user registration.
type Registerer interface {
	Register(ctx context.Context, username, password string) (*string, error)
}

// Loginer defines the interface for user authentication (login).
type Loginer interface {
	Login(ctx context.Context, username, password string) (*string, error)
}

// AuthServer implements the AuthServiceServer gRPC interface.
type AuthServer struct {
	pb.UnimplementedAuthServiceServer

	registerer Registerer
	loginer    Loginer
}

// NewAuthServer creates a new gRPC AuthServer with the provided registerer and loginer implementations.
func NewAuthServer(reg Registerer, login Loginer) *AuthServer {
	return &AuthServer{
		registerer: reg,
		loginer:    login,
	}
}

// Register handles user registration via gRPC.
func (s *AuthServer) Register(ctx context.Context, req *pb.AuthRequest) (*pb.AuthResponse, error) {
	token, err := s.registerer.Register(ctx, req.GetUsername(), req.GetPassword())
	if err != nil {
		return nil, err
	}

	var tokenStr string
	if token != nil {
		tokenStr = *token
	}

	return &pb.AuthResponse{Token: tokenStr}, nil
}

// Login handles user login via gRPC.
func (s *AuthServer) Login(ctx context.Context, req *pb.AuthRequest) (*pb.AuthResponse, error) {
	token, err := s.loginer.Login(ctx, req.GetUsername(), req.GetPassword())
	if err != nil {
		return nil, err
	}

	var tokenStr string
	if token != nil {
		tokenStr = *token
	}

	return &pb.AuthResponse{Token: tokenStr}, nil
}
