package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/sbilibin2017/gophkeeper/internal/services"
)

// Registerer defines interface for user registration.
type Registerer interface {
	// Register registers a new user, returns error if any.
	Register(ctx context.Context, username, password string) error
}

// Authenticator defines interface for user authentication (login).
type Authenticator interface {
	// Authenticate authenticates a user, returns error if any.
	Authenticate(ctx context.Context, username, password string) error
}

// JWTGenerator generates JWT tokens for users.
type JWTGenerator interface {
	Generate(username string) (string, error)
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

// NewRegisterHandler returns an HTTP handler for registering a new user.
// It accepts JSON body with username and password,
// creates a user, generates a JWT token, and returns it in Authorization header.
//
// @Summary Register a new user
// @Description Registers a user with username and password, returns JWT token
// @Tags auth
// @Accept json
// @Produce json
// @Param registerRequest body RegisterRequest true "Register request payload"
// @Success 200 {string} string "JWT token returned in Authorization header"
// @Failure 400 {string} string "invalid request body"
// @Failure 409 {string} string "user already exists"
// @Failure 500 {string} string "internal server error"
// @Router /register [post]
func NewRegisterHandler(auth Registerer, jwtGen JWTGenerator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req RegisterRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		// Register user
		err := auth.Register(r.Context(), req.Username, req.Password)
		if err != nil {
			switch {
			case errors.Is(err, services.ErrUserAlreadyExists):
				http.Error(w, err.Error(), http.StatusConflict)
			default:
				http.Error(w, "internal server error", http.StatusInternalServerError)
			}
			return
		}

		// Generate JWT token after successful registration
		token, err := jwtGen.Generate(req.Username)
		if err != nil {
			http.Error(w, "failed to generate token", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Authorization", "Bearer "+token)
		w.WriteHeader(http.StatusOK)
	}
}

// NewLoginHandler returns an HTTP handler for authenticating a user.
// It accepts JSON body with username and password,
// authenticates the user, generates a JWT token, and returns it in Authorization header.
//
// @Summary Authenticate a user (login)
// @Description Authenticates user and returns JWT token
// @Tags auth
// @Accept json
// @Produce json
// @Param loginRequest body LoginRequest true "Login request payload"
// @Success 200 {string} string "JWT token returned in Authorization header"
// @Failure 400 {string} string "invalid request body"
// @Failure 401 {string} string "invalid username or password"
// @Failure 500 {string} string "internal server error"
// @Router /login [post]
func NewLoginHandler(auth Authenticator, jwtGen JWTGenerator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req LoginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		// Authenticate user
		err := auth.Authenticate(r.Context(), req.Username, req.Password)
		if err != nil {
			switch {
			case errors.Is(err, services.ErrInvalidData):
				http.Error(w, err.Error(), http.StatusUnauthorized)
			default:
				http.Error(w, "internal server error", http.StatusInternalServerError)
			}
			return
		}

		// Generate JWT token after successful authentication
		token, err := jwtGen.Generate(req.Username)
		if err != nil {
			http.Error(w, "failed to generate token", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Authorization", "Bearer "+token)
		w.WriteHeader(http.StatusOK)
	}
}
