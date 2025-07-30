package grpc

import (
	"context"
	"errors"
	"strings"

	"github.com/sbilibin2017/gophkeeper/internal/models"
	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc"

	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// SecretWriter defines the interface for writing secrets to storage.
type SecretWriter interface {
	// Save stores a secret for a given user.
	Save(
		ctx context.Context,
		username string,
		secretName string,
		secretType string,
		ciphertext []byte,
		aesKeyEnc []byte,
	) error
}

// SecretReader defines the interface for reading secrets from storage.
type SecretReader interface {
	// Get retrieves a secret by type and name for a given user.
	Get(
		ctx context.Context,
		username string,
		secretType string,
		secretName string,
	) (*models.Secret, error)

	// List returns all secrets for a given user.
	List(
		ctx context.Context,
		username string,
	) ([]*models.Secret, error)
}

// JWTParser defines the interface for parsing JWT tokens.
type JWTParser interface {
	// Parse validates the token and returns the associated username.
	Parse(token string) (username string, err error)
}

// SecretWriteServer implements the SecretWriteService gRPC interface.
type SecretWriteServer struct {
	pb.UnimplementedSecretWriteServiceServer

	writer SecretWriter
	parser JWTParser
}

// NewSecretWriteServer creates a new SecretWriteServer instance.
//
// writer is the storage interface to save secrets.
// parser is used to parse and validate JWT tokens.
func NewSecretWriteServer(writer SecretWriter, parser JWTParser) *SecretWriteServer {
	return &SecretWriteServer{
		writer: writer,
		parser: parser,
	}
}

// Save handles saving a secret via gRPC.
//
// It extracts and validates the JWT token from gRPC metadata,
// extracts the username from the token,
// and saves the secret for the authenticated user.
func (s *SecretWriteServer) Save(ctx context.Context, req *pb.SecretSaveRequest) (*emptypb.Empty, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, errors.New("missing metadata in context")
	}

	authHeaders := md.Get("authorization")
	if len(authHeaders) == 0 {
		return nil, errors.New("missing authorization token")
	}

	authHeader := authHeaders[0]
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return nil, errors.New("invalid authorization token format")
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")

	username, err := s.parser.Parse(token)
	if err != nil {
		return nil, err
	}

	if err := s.writer.Save(ctx, username, req.GetSecretName(), req.GetSecretType(), req.GetCiphertext(), req.GetAesKeyEnc()); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

// SecretReadServer implements the SecretReadService gRPC interface.
type SecretReadServer struct {
	pb.UnimplementedSecretReadServiceServer

	reader SecretReader
	parser JWTParser
}

// NewSecretReadServer creates a new SecretReadServer instance.
//
// reader is the storage interface to read secrets.
// parser is used to parse and validate JWT tokens.
func NewSecretReadServer(reader SecretReader, parser JWTParser) *SecretReadServer {
	return &SecretReadServer{
		reader: reader,
		parser: parser,
	}
}

// Get handles fetching a single secret via gRPC.
//
// It extracts and validates the JWT token from gRPC metadata,
// extracts the username from the token,
// and returns the secret associated with the authenticated user.
func (s *SecretReadServer) Get(ctx context.Context, req *pb.SecretGetRequest) (*pb.Secret, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, errors.New("missing metadata in context")
	}

	authHeaders := md.Get("authorization")
	if len(authHeaders) == 0 {
		return nil, errors.New("missing authorization token")
	}

	authHeader := authHeaders[0]
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return nil, errors.New("invalid authorization token format")
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")

	username, err := s.parser.Parse(token)
	if err != nil {
		return nil, err
	}

	secret, err := s.reader.Get(ctx, username, req.GetSecretType(), req.GetSecretName())
	if err != nil {
		return nil, err
	}

	return &pb.Secret{
		SecretName:  secret.SecretName,
		SecretType:  secret.SecretType,
		SecretOwner: secret.SecretOwner,
		Ciphertext:  secret.Ciphertext,
		AesKeyEnc:   secret.AESKeyEnc,
		CreatedAt:   timestamppb.New(secret.CreatedAt),
		UpdatedAt:   timestamppb.New(secret.UpdatedAt),
	}, nil
}

// List streams all secrets for the authenticated user via gRPC.
//
// It extracts and validates the JWT token from gRPC metadata,
// extracts the username from the token,
// then streams all secrets associated with the user.
func (s *SecretReadServer) List(empty *emptypb.Empty, stream pb.SecretReadService_ListServer) error {
	ctx := stream.Context()

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return errors.New("missing metadata in context")
	}

	authHeaders := md.Get("authorization")
	if len(authHeaders) == 0 {
		return errors.New("missing authorization token")
	}

	authHeader := authHeaders[0]
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return errors.New("invalid authorization token format")
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")

	username, err := s.parser.Parse(token)
	if err != nil {
		return err
	}

	secrets, err := s.reader.List(ctx, username)
	if err != nil {
		return err
	}

	for _, secret := range secrets {
		if err := stream.Send(&pb.Secret{
			SecretName:  secret.SecretName,
			SecretType:  secret.SecretType,
			SecretOwner: secret.SecretOwner,
			Ciphertext:  secret.Ciphertext,
			AesKeyEnc:   secret.AESKeyEnc,
			CreatedAt:   timestamppb.New(secret.CreatedAt),
			UpdatedAt:   timestamppb.New(secret.UpdatedAt),
		}); err != nil {
			return err
		}
	}

	return nil
}
