package services

import (
	"context"
	"errors"

	"github.com/sbilibin2017/gophkeeper/internal/models"
	"golang.org/x/crypto/bcrypt"
)

type UserReader interface {
	Get(ctx context.Context, username string) (*models.User, error)
}

type UserWriter interface {
	Save(ctx context.Context, username, passwordHash string) error
}

type AuthService struct {
	writer UserWriter
	reader UserReader
}

func (svc *AuthService) Register(
	ctx context.Context,
	username string,
	password string,
) error {
	existingUser, err := svc.reader.Get(ctx, username)
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

	return svc.writer.Save(ctx, username, string(hashedPassword))
}

func (svc *AuthService) Login(
	ctx context.Context,
	username string,
	password string,
) error {
	user, err := svc.reader.Get(ctx, username)
	if err != nil {
		return err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return errors.New("invalid username or password")
	}

	return nil
}

var (
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrInvalidInput      = errors.New("username and password must be provided")
)
