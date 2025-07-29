package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/sbilibin2017/gophkeeper/internal/models"
	"golang.org/x/crypto/bcrypt"
)

// UserSaver defines the interface for persisting a new user with hashed password.
type UserSaver interface {
	Save(ctx context.Context, username, passwordHash string) error
}

// UserGetter defines the interface for retrieving a user by username.
type UserGetter interface {
	Get(ctx context.Context, username string) (*models.User, error)
}

// JWTGenerator defines the interface for generating JWT tokens for authenticated users.
type JWTGenerator interface {
	Generate(username string) (string, error)
}

var (
	// errUserAlreadyExists is returned when a user attempts to register with an existing username.
	errUserAlreadyExists = errors.New("user already exists")

	// errInvalidLogin is returned when login credentials are invalid.
	errInvalidLogin = errors.New("invalid username or password")
)

// NewRegisterHandler returns an HTTP handler function for user registration.
// It expects a JSON body with "username" and "password" fields.
// On success, it returns a JWT token in the "Authorization" header.

// RegisterRequest represents the expected request body for user registration.
// swagger:model RegisterRequest
type RegisterRequest struct {
	// Username for the new user
	// example: johndoe
	Username string `json:"username"`
	// Password for the new user
	// example: secret123
	Password string `json:"password"`
}

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
func NewRegisterHandler(
	userGetter UserGetter,
	userSaver UserSaver,
	jwtGenerator JWTGenerator,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		var req RegisterRequest

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		// Check if user already exists
		existingUser, err := userGetter.Get(ctx, req.Username)
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		if existingUser != nil {
			http.Error(w, errUserAlreadyExists.Error(), http.StatusConflict)
			return
		}

		// Hash password
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			http.Error(w, "could not hash password", http.StatusInternalServerError)
			return
		}

		// Save user
		if err := userSaver.Save(ctx, req.Username, string(hashedPassword)); err != nil {
			http.Error(w, "could not save user", http.StatusInternalServerError)
			return
		}

		// Generate JWT token
		token, err := jwtGenerator.Generate(req.Username)
		if err != nil {
			http.Error(w, "failed to generate token", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Authorization", "Bearer "+token)
		w.WriteHeader(http.StatusOK)
	}
}

// NewLoginHandler returns an HTTP handler function for user authentication (login).
// It expects a JSON body with "username" and "password" fields.
// On success, it returns a JWT token in the "Authorization" header.

// LoginRequest represents the expected request body for user login.
// swagger:model LoginRequest
type LoginRequest struct {
	// Username of the user
	// example: johndoe
	Username string `json:"username"`
	// Password of the user
	// example: secret123
	Password string `json:"password"`
}

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
func NewLoginHandler(
	userGetter UserGetter,
	jwtGenerator JWTGenerator,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		var req LoginRequest

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		// Fetch user from storage
		user, err := userGetter.Get(ctx, req.Username)
		if err != nil {
			http.Error(w, errInvalidLogin.Error(), http.StatusUnauthorized)
			return
		}

		// Compare password with stored hash
		err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
		if err != nil {
			http.Error(w, errInvalidLogin.Error(), http.StatusUnauthorized)
			return
		}

		// Generate JWT token
		token, err := jwtGenerator.Generate(req.Username)
		if err != nil {
			http.Error(w, "failed to generate token", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Authorization", "Bearer "+token)
		w.WriteHeader(http.StatusOK)
	}
}
