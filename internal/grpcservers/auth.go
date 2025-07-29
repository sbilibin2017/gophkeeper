package grpcservers

import (
	"context"

	"github.com/sbilibin2017/gophkeeper/internal/models"
	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// UserSaver defines the interface for persisting a new user with a hashed password.
type UserSaver interface {
	// Save stores a new user with a hashed password.
	Save(ctx context.Context, username, passwordHash string) error
}

// UserGetter defines the interface for retrieving a user by username.
type UserGetter interface {
	// Get retrieves a user by their username.
	Get(ctx context.Context, username string) (*models.User, error)
}

// JWTGenerator defines the interface for generating JWT tokens for authenticated users.
type JWTGenerator interface {
	// Generate creates a JWT token for the given username.
	Generate(username string) (string, error)
}

// AuthServer implements the pb.AuthServiceServer gRPC interface for user authentication.
type AuthServer struct {
	pb.UnimplementedAuthServiceServer

	userSaver    UserSaver
	userGetter   UserGetter
	jwtGenerator JWTGenerator
}

// NewAuthServer creates a new AuthServer instance with the provided dependencies.
//
// It accepts a UserSaver for storing users, a UserGetter for retrieving users,
// and a JWTGenerator for issuing authentication tokens.
func NewAuthServer(userSaver UserSaver, userGetter UserGetter, jwtGen JWTGenerator) *AuthServer {
	return &AuthServer{
		userSaver:    userSaver,
		userGetter:   userGetter,
		jwtGenerator: jwtGen,
	}
}

// Register registers a new user with a hashed password and returns a JWT token.
//
// It returns a gRPC AlreadyExists error if the user already exists,
// or Internal error if hashing or storage fails.
func (s *AuthServer) Register(ctx context.Context, req *pb.AuthRequest) (*pb.AuthResponse, error) {
	username := req.GetUsername()
	password := req.GetPassword()

	existingUser, err := s.userGetter.Get(ctx, username)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "error checking existing user: %v", err)
	}
	if existingUser != nil {
		return nil, status.Error(codes.AlreadyExists, "user already exists")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not hash password: %v", err)
	}

	if err := s.userSaver.Save(ctx, username, string(hashedPassword)); err != nil {
		return nil, status.Errorf(codes.Internal, "could not save user: %v", err)
	}

	token, err := s.jwtGenerator.Generate(username)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to generate token: %v", err)
	}

	return &pb.AuthResponse{Token: token}, nil
}

// Login authenticates an existing user and returns a JWT token.
//
// It returns a gRPC Unauthenticated error if the username or password is incorrect,
// or Internal error if token generation fails.
func (s *AuthServer) Login(ctx context.Context, req *pb.AuthRequest) (*pb.AuthResponse, error) {
	username := req.GetUsername()
	password := req.GetPassword()

	user, err := s.userGetter.Get(ctx, username)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid username or password")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid username or password")
	}

	token, err := s.jwtGenerator.Generate(username)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to generate token: %v", err)
	}

	return &pb.AuthResponse{Token: token}, nil
}
