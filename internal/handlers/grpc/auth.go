package grpc

import (
	"context"

	"github.com/sbilibin2017/gophkeeper/internal/services"
	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Registerer defines interface for user registration.
type AuthService interface {
	Register(ctx context.Context, username, password string) error
	Authenticate(ctx context.Context, username, password string) error
}

// JWTGenerator generates JWT tokens for users.
type JWTGenerator interface {
	Generate(username string) (string, error)
}

// AuthServer implements the gRPC AuthService using the above interfaces.
type AuthServer struct {
	pb.UnimplementedAuthServiceServer

	svc          AuthService
	jwtGenerator JWTGenerator
}

// NewAuthServer creates a new AuthServer instance with the provided interfaces.
func NewAuthServer(
	svc AuthService,
	jwtGen JWTGenerator,
) *AuthServer {
	return &AuthServer{
		svc:          svc,
		jwtGenerator: jwtGen,
	}
}

// Register implements user registration via gRPC.
func (s *AuthServer) Register(ctx context.Context, req *pb.AuthRequest) (*pb.AuthResponse, error) {
	err := s.svc.Register(ctx, req.GetUsername(), req.GetPassword())
	if err != nil {
		switch {
		case err == services.ErrUserAlreadyExists:
			return nil, status.Error(codes.AlreadyExists, err.Error())
		default:
			return nil, status.Error(codes.Internal, "internal server error")
		}
	}

	token, err := s.jwtGenerator.Generate(req.GetUsername())
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to generate token")
	}

	return &pb.AuthResponse{Token: token}, nil
}

// Login implements user authentication via gRPC.
func (s *AuthServer) Login(ctx context.Context, req *pb.AuthRequest) (*pb.AuthResponse, error) {
	err := s.svc.Authenticate(ctx, req.GetUsername(), req.GetPassword())
	if err != nil {
		switch {
		case err == services.ErrInvalidData:
			return nil, status.Error(codes.Unauthenticated, err.Error())
		default:
			return nil, status.Error(codes.Internal, "internal server error")
		}
	}

	token, err := s.jwtGenerator.Generate(req.GetUsername())
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to generate token")
	}

	return &pb.AuthResponse{Token: token}, nil
}
