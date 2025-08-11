package services

import (
	"context"
	"errors"

	"github.com/sbilibin2017/gophkeeper/internal/models"
)

// UserSaver defines an interface to save user data.
type UserSaver interface {
	Save(ctx context.Context, user *models.UserDB) error
}

// UserGetter defines an interface to retrieve user data by username.
type UserGetter interface {
	Get(ctx context.Context, username string) (*models.UserDB, error)
}

// Tokener defines an interface to generate JWT tokens for a username.
type Tokener interface {
	Generate(username string) (string, error)
}

// Hasher defines an interface for hashing and comparing hashed values.
type Hasher interface {
	Hash(value []byte) ([]byte, error)
	Compare(hashedValue []byte, value []byte) error
}

var (
	// ErrUserAlreadyExists is returned when attempting to register a username that already exists.
	ErrUserAlreadyExists = errors.New("user already exists")

	// ErrInvalidData is returned when provided credentials are invalid.
	ErrInvalidData = errors.New("invalid username or password")
)

// AuthService provides user registration and authentication services.
type AuthService struct {
	getter  UserGetter
	saver   UserSaver
	hasher  Hasher
	tokener Tokener
}

// NewAuthService creates a new AuthService with given dependencies.
func NewAuthService(
	getter UserGetter,
	saver UserSaver,
	hasher Hasher,
	tokener Tokener,
) *AuthService {
	return &AuthService{
		getter:  getter,
		saver:   saver,
		hasher:  hasher,
		tokener: tokener,
	}
}

// Register creates a new user with the given username and password.
// It hashes the password before saving the user.
// Returns ErrUserAlreadyExists if the username is taken.
func (s *AuthService) Register(
	ctx context.Context,
	user *models.User,
) (string, error) {
	existingUser, err := s.getter.Get(ctx, user.Username)
	if err != nil {
		return "", err
	}
	if existingUser != nil {
		return "", ErrUserAlreadyExists
	}

	hashedPassword, err := s.hasher.Hash([]byte(user.Password))
	if err != nil {
		return "", err
	}

	row := &models.UserDB{
		Username:     user.Username,
		PasswordHash: string(hashedPassword),
	}

	if err := s.saver.Save(ctx, row); err != nil {
		return "", err
	}

	token, err := s.tokener.Generate(user.Username)
	if err != nil {
		return "", err
	}

	return token, nil
}

// Authenticate verifies user credentials and returns a JWT token upon success.
// Returns ErrInvalidData if the username does not exist or password is incorrect.
func (s *AuthService) Authenticate(
	ctx context.Context,
	user *models.User,
) (string, error) {
	if user == nil {
		return "", ErrInvalidData
	}
	row, err := s.getter.Get(ctx, user.Username)
	if err != nil {
		return "", err
	}
	if row == nil {
		return "", ErrInvalidData
	}

	if err := s.hasher.Compare([]byte(row.PasswordHash), []byte(user.Password)); err != nil {
		return "", ErrInvalidData
	}

	token, err := s.tokener.Generate(user.Username)
	if err != nil {
		return "", err
	}

	return token, nil
}
