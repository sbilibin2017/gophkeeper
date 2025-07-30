package facades

import (
	"context"
	"fmt"
	"io"

	"github.com/go-resty/resty/v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	emptypb "google.golang.org/protobuf/types/known/emptypb"

	"github.com/sbilibin2017/gophkeeper/internal/models"
	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc"
)

//
// HTTP Facades
//

type SecretWriterHTTP struct {
	client *resty.Client
}

func NewSecretWriterHTTP(client *resty.Client) *SecretWriterHTTP {
	return &SecretWriterHTTP{client: client}
}

// Save sends a secret to be stored via HTTP.
func (w *SecretWriterHTTP) Save(
	ctx context.Context,
	secretOwner string,
	secretName string,
	secretType string,
	ciphertext []byte,
	aesKeyEnc []byte,
) error {
	secret := &models.Secret{
		SecretOwner: secretOwner,
		SecretName:  secretName,
		SecretType:  secretType,
		Ciphertext:  ciphertext,
		AESKeyEnc:   aesKeyEnc,
	}

	resp, err := w.client.R().
		SetContext(ctx).
		SetAuthToken(secretOwner).
		SetBody(secret).
		Post("/save")
	if err != nil {
		return fmt.Errorf("http save request failed: %w", err)
	}
	if resp.IsError() {
		return fmt.Errorf("http error status %d, body: %s", resp.StatusCode(), resp.String())
	}
	return nil
}

type SecretReaderHTTP struct {
	client *resty.Client
}

func NewSecretReaderHTTP(client *resty.Client) *SecretReaderHTTP {
	return &SecretReaderHTTP{client: client}
}

// Get fetches a secret by owner, type, and name via HTTP.
func (r *SecretReaderHTTP) Get(
	ctx context.Context,
	secretOwner string,
	secretType string,
	secretName string,
) (*models.Secret, error) {
	var secret models.Secret
	resp, err := r.client.R().
		SetContext(ctx).
		SetResult(&secret).
		SetAuthToken(secretOwner).
		SetPathParam("secretType", secretType).
		SetPathParam("secretName", secretName).
		Get("/get/{secretType}/{secretName}")
	if err != nil {
		return nil, fmt.Errorf("http get request failed: %w", err)
	}
	if resp.IsError() {
		return nil, fmt.Errorf("http error status %d, body: %s", resp.StatusCode(), resp.String())
	}
	return &secret, nil
}

// List fetches all secrets for a given owner via HTTP.
func (r *SecretReaderHTTP) List(
	ctx context.Context,
	secretOwner string,
) ([]*models.Secret, error) {
	var secrets []*models.Secret
	resp, err := r.client.R().
		SetContext(ctx).
		SetResult(&secrets).
		SetAuthToken(secretOwner).
		Get("/list")
	if err != nil {
		return nil, fmt.Errorf("http list request failed: %w", err)
	}
	if resp.IsError() {
		return nil, fmt.Errorf("http error status %d, body: %s", resp.StatusCode(), resp.String())
	}
	return secrets, nil
}

//
// gRPC Facades
//

type SecretWriterGRPC struct {
	client pb.SecretWriteServiceClient
}

func NewSecretWriterGRPC(conn *grpc.ClientConn) *SecretWriterGRPC {
	return &SecretWriterGRPC{
		client: pb.NewSecretWriteServiceClient(conn),
	}
}

// Save sends a secret to be stored via gRPC.
func (w *SecretWriterGRPC) Save(
	ctx context.Context,
	secretOwner string,
	secretName string,
	secretType string,
	ciphertext []byte,
	aesKeyEnc []byte,
) error {
	// Inject secretOwner as metadata in the outgoing context
	ctx = metadata.NewOutgoingContext(ctx, metadata.Pairs("token", secretOwner))

	req := &pb.SecretSaveRequest{
		SecretName: secretName,
		SecretType: secretType,
		Ciphertext: ciphertext,
		AesKeyEnc:  aesKeyEnc,
	}

	_, err := w.client.Save(ctx, req)
	if err != nil {
		return fmt.Errorf("gRPC save failed: %w", err)
	}
	return nil
}

type SecretReaderGRPC struct {
	client pb.SecretReadServiceClient
}

func NewSecretReaderGRPC(conn *grpc.ClientConn) *SecretReaderGRPC {
	return &SecretReaderGRPC{
		client: pb.NewSecretReadServiceClient(conn),
	}
}

// Get fetches a secret by owner, type, and name via gRPC.
func (r *SecretReaderGRPC) Get(
	ctx context.Context,
	secretOwner string,
	secretType string,
	secretName string,
) (*models.Secret, error) {
	ctx = metadata.NewOutgoingContext(ctx, metadata.Pairs("token", secretOwner))

	req := &pb.SecretGetRequest{
		SecretName: secretName,
		SecretType: secretType,
	}

	resp, err := r.client.Get(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("gRPC Get failed: %w", err)
	}

	return &models.Secret{
		SecretOwner: resp.SecretOwner,
		SecretName:  resp.SecretName,
		SecretType:  resp.SecretType,
		Ciphertext:  resp.Ciphertext,
		AESKeyEnc:   resp.AesKeyEnc,
		CreatedAt:   resp.CreatedAt.AsTime(),
		UpdatedAt:   resp.UpdatedAt.AsTime(),
	}, nil
}

// List fetches all secrets for a given owner via gRPC.
func (r *SecretReaderGRPC) List(
	ctx context.Context,
	secretOwner string,
) ([]*models.Secret, error) {
	ctx = metadata.NewOutgoingContext(ctx, metadata.Pairs("token", secretOwner))

	stream, err := r.client.List(ctx, &emptypb.Empty{})
	if err != nil {
		return nil, fmt.Errorf("gRPC List stream start failed: %w", err)
	}

	var secrets []*models.Secret
	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("gRPC List stream receive failed: %w", err)
		}

		secrets = append(secrets, &models.Secret{
			SecretOwner: resp.SecretOwner,
			SecretName:  resp.SecretName,
			SecretType:  resp.SecretType,
			Ciphertext:  resp.Ciphertext,
			AESKeyEnc:   resp.AesKeyEnc,
			CreatedAt:   resp.CreatedAt.AsTime(),
			UpdatedAt:   resp.UpdatedAt.AsTime(),
		})
	}

	return secrets, nil
}
