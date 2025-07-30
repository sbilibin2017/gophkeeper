package services

import (
	"context"
	"errors"

	"golang.org/x/crypto/bcrypt"

	"github.com/sbilibin2017/gophkeeper/internal/models"
)

// Dependencies needed by the service
type UserSaver interface {
	Save(ctx context.Context, username, passwordHash string) error
}

type UserGetter interface {
	Get(ctx context.Context, username string) (*models.User, error)
}

type JWTGenerator interface {
	Generate(username string) (string, error)
}

var (
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrInvalidData       = errors.New("invalid username or password")
)

type AuthService struct {
	users UserGetter
	saver UserSaver
}

func NewAuthService(users UserGetter, saver UserSaver) *AuthService {
	return &AuthService{
		users: users,
		saver: saver,
	}
}

// Register a new user and return JWT token
func (s *AuthService) Register(ctx context.Context, username, password string) error {
	existingUser, err := s.users.Get(ctx, username)
	if err != nil {
		return err
	}
	if existingUser != nil {
		return ErrUserAlreadyExists
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	if err := s.saver.Save(ctx, username, string(hashedPassword)); err != nil {
		return err
	}

	return nil
}

// Authenticate verifies credentials and returns JWT token
func (s *AuthService) Authenticate(ctx context.Context, username, password string) error {
	user, err := s.users.Get(ctx, username)
	if err != nil || user == nil {
		return ErrInvalidData
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return ErrInvalidData
	}

	return nil
}
