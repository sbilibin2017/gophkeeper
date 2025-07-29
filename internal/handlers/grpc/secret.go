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
	// Save stores a secret for the given user.
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
	// Get retrieves a secret by type and name for the given user.
	Get(
		ctx context.Context,
		username string,
		secretType string,
		secretName string,
	) (*models.Secret, error)

	// List returns all secrets for the given user.
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

// SecretWriteServiceServer implements SecretWriteService gRPC interface.
type SecretWriteServiceServer struct {
	pb.UnimplementedSecretWriteServiceServer

	writer    SecretWriter
	jwtParser JWTParser
}

// NewSecretWriteServiceServer creates a new SecretWriteServiceServer.
func NewSecretWriteServiceServer(writer SecretWriter, jwtParser JWTParser) *SecretWriteServiceServer {
	return &SecretWriteServiceServer{
		writer:    writer,
		jwtParser: jwtParser,
	}
}

// Save handles saving a secret via gRPC.
//
// It extracts and validates the JWT token from metadata,
// extracts the username from the token,
// then saves the secret associated with the user.
func (s *SecretWriteServiceServer) Save(ctx context.Context, req *pb.SecretSaveRequest) (*emptypb.Empty, error) {
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

	username, err := s.jwtParser.Parse(token)
	if err != nil {
		return nil, err
	}

	if err := s.writer.Save(ctx, username, req.GetSecretName(), req.GetSecretType(), req.GetCiphertext(), req.GetAesKeyEnc()); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

// SecretReadServiceServer implements SecretReadService gRPC interface.
type SecretReadServiceServer struct {
	pb.UnimplementedSecretReadServiceServer

	reader    SecretReader
	jwtParser JWTParser
}

// NewSecretReadServiceServer creates a new SecretReadServiceServer.
func NewSecretReadServiceServer(reader SecretReader, jwtParser JWTParser) *SecretReadServiceServer {
	return &SecretReadServiceServer{
		reader:    reader,
		jwtParser: jwtParser,
	}
}

// Get handles fetching a single secret via gRPC.
//
// It extracts and validates the JWT token from metadata,
// extracts the username from the token,
// then fetches the secret associated with the user.
func (s *SecretReadServiceServer) Get(ctx context.Context, req *pb.SecretGetRequest) (*pb.Secret, error) {
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

	username, err := s.jwtParser.Parse(token)
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

// List streams all secrets for the authenticated user.
//
// It extracts and validates the JWT token from metadata,
// extracts the username from the token,
// then streams all secrets associated with the user.
func (s *SecretReadServiceServer) List(empty *emptypb.Empty, stream pb.SecretReadService_ListServer) error {
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

	username, err := s.jwtParser.Parse(token)
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
