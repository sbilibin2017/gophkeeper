package services

import (
	"context"

	"github.com/sbilibin2017/gophkeeper/internal/models"
)

// SecretWriter defines the interface that the write service depends on.
type SecretWriter interface {
	Save(
		ctx context.Context,
		username, secretName, secretType string,
		ciphertext, aesKeyEnc []byte,
	) error
}

// SecretWriteService provides methods for writing secrets.
type SecretWriteService struct {
	writer SecretWriter
}

// NewSecretWriteService creates a new instance of SecretWriteService.
func NewSecretWriteService(writer SecretWriter) *SecretWriteService {
	return &SecretWriteService{writer: writer}
}

// Save stores a new secret.
func (s *SecretWriteService) Save(
	ctx context.Context,
	username, secretName, secretType string,
	ciphertext, aesKeyEnc []byte,
) error {
	return s.writer.Save(ctx, username, secretName, secretType, ciphertext, aesKeyEnc)
}

// SecretReader defines the interface that the read service depends on.
type SecretReader interface {
	Get(ctx context.Context, username, typ, name string) (*models.Secret, error)
	List(ctx context.Context, username string) ([]*models.Secret, error)
}

// SecretReadService provides methods for reading secrets using a JWT token.
type SecretReadService struct {
	reader SecretReader
}

// NewSecretReadService creates a new instance of SecretReadService.
func NewSecretReadService(reader SecretReader) *SecretReadService {
	return &SecretReadService{
		reader: reader,
	}
}

// Get parses token and returns a secret by type and name.
func (s *SecretReadService) Get(
	ctx context.Context,
	username, secretType, secretName string,
) (*models.Secret, error) {
	secret, err := s.reader.Get(ctx, username, secretType, secretName)
	if err != nil {
		return nil, err
	}
	return secret, nil
}

// List parses token and returns all secrets for the user.
func (s *SecretReadService) List(
	ctx context.Context,
	username string,
) ([]*models.Secret, error) {
	return s.reader.List(ctx, username)
}
