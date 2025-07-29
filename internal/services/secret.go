package services

import (
	"context"
	"errors"

	"github.com/sbilibin2017/gophkeeper/internal/models"
)

// SecretWriter defines write operations for secrets.
type SecretWriter interface {
	Save(
		ctx context.Context,
		secretOwner string,
		secretName string,
		secretType string,
		ciphertext []byte,
		aesKeyEnc []byte,
	) error
}

// SecretReader defines read operations for secrets.
type SecretReader interface {
	Get(ctx context.Context, secretOwner, secretType, secretName string) (*models.Secret, error)
	List(ctx context.Context, secretOwner string) ([]*models.Secret, error)
}

// JWTParser defines the interface for parsing JWT tokens.
type JWTParser interface {
	Parse(tokenStr string) (string, error)
}

type SecretWriteService struct {
	secretWriter SecretWriter
	jwtParser    JWTParser
}

func NewSecretWriteService(writer SecretWriter, jwtParser JWTParser) *SecretWriteService {
	return &SecretWriteService{
		secretWriter: writer,
		jwtParser:    jwtParser,
	}
}

func (s *SecretWriteService) Save(
	ctx context.Context,
	token string,
	secretName string,
	secretType string,
	ciphertext []byte,
	aesKeyEnc []byte,
) error {
	username, err := s.jwtParser.Parse(token)
	if err != nil {
		return ErrInvalidToken
	}

	return s.secretWriter.Save(ctx, username, secretName, secretType, ciphertext, aesKeyEnc)
}

type SecretReadService struct {
	secretReader SecretReader
	jwtParser    JWTParser
}

func NewSecretReadService(reader SecretReader, jwtParser JWTParser) *SecretReadService {
	return &SecretReadService{
		secretReader: reader,
		jwtParser:    jwtParser,
	}
}

func (s *SecretReadService) Get(
	ctx context.Context,
	token string,
	secretType string,
	secretName string,
) (*models.Secret, error) {
	username, err := s.jwtParser.Parse(token)
	if err != nil {
		return nil, ErrInvalidToken
	}

	return s.secretReader.Get(ctx, username, secretType, secretName)
}

func (s *SecretReadService) List(
	ctx context.Context,
	token string,
) ([]*models.Secret, error) {
	username, err := s.jwtParser.Parse(token)
	if err != nil {
		return nil, ErrInvalidToken
	}

	return s.secretReader.List(ctx, username)
}

var (
	ErrInvalidToken = errors.New("invalid token")
)
