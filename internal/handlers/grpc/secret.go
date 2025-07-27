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

// SecretWriter defines write operations for secrets.
type SecretWriter interface {
	Save(ctx context.Context, secretOwner, secretName, secretType string, ciphertext, aesKeyEnc []byte) error
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

// SecretWriteServiceServer implements SecretWriteService gRPC interface.
type SecretWriteServiceServer struct {
	pb.UnimplementedSecretWriteServiceServer

	writer    SecretWriter
	jwtParser JWTParser
}

func NewSecretWriteServiceServer(writer SecretWriter, jwtParser JWTParser) *SecretWriteServiceServer {
	return &SecretWriteServiceServer{
		writer:    writer,
		jwtParser: jwtParser,
	}
}

// Save implements the Save RPC for SecretWriteService.
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
	owner, err := s.jwtParser.Parse(token)
	if err != nil {
		return nil, err
	}

	err = s.writer.Save(ctx, owner, req.GetSecretName(), req.GetSecretType(), req.GetCiphertext(), req.GetAesKeyEnc())
	if err != nil {
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

func NewSecretReadServiceServer(reader SecretReader, jwtParser JWTParser) *SecretReadServiceServer {
	return &SecretReadServiceServer{
		reader:    reader,
		jwtParser: jwtParser,
	}
}

// Get implements the Get RPC for SecretReadService.
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
	owner, err := s.jwtParser.Parse(token)
	if err != nil {
		return nil, err
	}

	secret, err := s.reader.Get(ctx, owner, req.GetSecretType(), req.GetSecretName())
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

// List implements the List RPC for SecretReadService.
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
	owner, err := s.jwtParser.Parse(token)
	if err != nil {
		return err
	}

	secrets, err := s.reader.List(ctx, owner)
	if err != nil {
		return err
	}

	for _, secret := range secrets {
		err := stream.Send(&pb.Secret{
			SecretName:  secret.SecretName,
			SecretType:  secret.SecretType,
			SecretOwner: secret.SecretOwner,
			Ciphertext:  secret.Ciphertext,
			AesKeyEnc:   secret.AESKeyEnc,
			CreatedAt:   timestamppb.New(secret.CreatedAt),
			UpdatedAt:   timestamppb.New(secret.UpdatedAt),
		})
		if err != nil {
			return err
		}
	}

	return nil
}
