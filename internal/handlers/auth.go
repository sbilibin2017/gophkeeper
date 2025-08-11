package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/sbilibin2017/gophkeeper/internal/models"
	"github.com/sbilibin2017/gophkeeper/internal/services"
	"github.com/sbilibin2017/gophkeeper/internal/validators"
	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// UserService defines interface for user service.
type UserService interface {
	// Register registers a new user and returns a JWT token or error.
	Register(ctx context.Context, user *models.User) (string, error)
	// Authenticate authenticates a user and returns a JWT token or error.
	Authenticate(ctx context.Context, user *models.User) (string, error)
}

// RegisterRequest represents the expected request body for user registration.
// swagger:model RegisterRequest
type RegisterRequest struct {
	// Username for the new user
	// example: johndoe
	Username string `json:"username" example:"johndoe"`
	// Password for the new user
	// example: secret123
	Password string `json:"password" example:"secret123"`
}

// LoginRequest represents the expected request body for user login.
// swagger:model LoginRequest
type LoginRequest struct {
	// Username of the user
	// example: johndoe
	Username string `json:"username" example:"johndoe"`
	// Password of the user
	// example: secret123
	Password string `json:"password" example:"secret123"`
}

// HTTPHandler handles HTTP requests for authentication.
type HTTPHandler struct {
	svc               UserService
	usernameValidator func(username string) error
	passwordValidator func(password string) error
}

// NewHTTPHandler creates a new HTTPHandler with the given UserService.
func NewHTTPHandler(
	svc UserService,
	usernameValidator func(username string) error,
	passwordValidator func(password string) error,
) *HTTPHandler {
	return &HTTPHandler{
		svc:               svc,
		usernameValidator: usernameValidator,
		passwordValidator: passwordValidator}
}

// Register handles user registration.
// @Summary Register a new user
// @Description Registers a user with username and password, returns JWT token in Authorization header
// @Tags auth
// @Accept json
// @Produce json
// @Param registerRequest body RegisterRequest true "Register request payload"
// @Success 200 {string} string "JWT token returned in Authorization header"
// @Failure 400 {string} string "invalid request body or validation failed"
// @Failure 409 {string} string "user already exists"
// @Failure 500 {string} string "internal server error"
// @Router /register [post]
func (h *HTTPHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Validate username
	if err := h.usernameValidator(req.Username); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Validate password
	if err := h.passwordValidator(req.Password); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	token, err := h.svc.Register(r.Context(), &models.User{
		Username: req.Username,
		Password: req.Password,
	})
	if err != nil {
		switch {
		case errors.Is(err, services.ErrUserAlreadyExists):
			w.WriteHeader(http.StatusConflict)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Authorization", "Bearer "+token)
	w.WriteHeader(http.StatusOK)
}

// Login handles user authentication.
// @Summary Authenticate a user (login)
// @Description Authenticates user and returns JWT token in Authorization header
// @Tags auth
// @Accept json
// @Produce json
// @Param loginRequest body LoginRequest true "Login request payload"
// @Success 200 {string} string "JWT token returned in Authorization header"
// @Failure 400 {string} string "invalid request body"
// @Failure 401 {string} string "invalid username or password"
// @Failure 500 {string} string "internal server error"
// @Router /login [post]
func (h *HTTPHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	token, err := h.svc.Authenticate(r.Context(), &models.User{
		Username: req.Username,
		Password: req.Password,
	})
	if err != nil {
		switch {
		case errors.Is(err, services.ErrInvalidData):
			w.WriteHeader(http.StatusUnauthorized)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Authorization", "Bearer "+token)
	w.WriteHeader(http.StatusOK)
}

// GRPCHandler handles gRPC authentication requests.
type GRPCHandler struct {
	pb.UnimplementedAuthServiceServer
	svc               UserService
	usernameValidator func(string) error
	passwordValidator func(string) error
}

// NewGRPCHandler creates a new GRPCHandler with injected validators.
func NewGRPCHandler(
	svc UserService,
	usernameValidator func(string) error,
	passwordValidator func(string) error,
) *GRPCHandler {
	if usernameValidator == nil {
		usernameValidator = validators.ValidateUsername
	}
	if passwordValidator == nil {
		passwordValidator = validators.ValidatePassword
	}
	return &GRPCHandler{
		svc:               svc,
		usernameValidator: usernameValidator,
		passwordValidator: passwordValidator,
	}
}

// Register processes a gRPC user registration request.
func (h *GRPCHandler) Register(ctx context.Context, req *pb.AuthRequest) (*pb.AuthResponse, error) {
	user := &models.User{
		Username: req.GetUsername(),
		Password: req.GetPassword(),
	}

	if err := h.usernameValidator(user.Username); err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid username")
	}
	if err := h.passwordValidator(user.Password); err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid password")
	}

	token, err := h.svc.Register(ctx, user)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrUserAlreadyExists):
			return nil, status.Error(codes.AlreadyExists, "user already exists")
		default:
			return nil, status.Error(codes.Internal, "internal server error")
		}
	}

	return &pb.AuthResponse{Token: token}, nil
}

// Login processes a gRPC user login request.
func (h *GRPCHandler) Login(ctx context.Context, req *pb.AuthRequest) (*pb.AuthResponse, error) {
	user := &models.User{
		Username: req.GetUsername(),
		Password: req.GetPassword(),
	}

	token, err := h.svc.Authenticate(ctx, user)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrInvalidData):
			return nil, status.Error(codes.Unauthenticated, "invalid username or password")
		default:
			return nil, status.Error(codes.Internal, "internal server error")
		}
	}

	return &pb.AuthResponse{Token: token}, nil
}
