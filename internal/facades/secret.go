package facades

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/go-resty/resty/v2"
	"github.com/sbilibin2017/gophkeeper/internal/models"
	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/emptypb"
)

// SecretWriteHTTPFacade handles saving secrets over HTTP.
type SecretWriteHTTPFacade struct {
	client *resty.Client
}

func NewSecretWriteHTTPFacade(client *resty.Client) *SecretWriteHTTPFacade {
	return &SecretWriteHTTPFacade{client: client}
}

// Save saves a secret using explicit parameters (no struct).
func (f *SecretWriteHTTPFacade) Save(
	ctx context.Context,
	secretOwner string,
	secretName string,
	secretType string,
	ciphertext []byte,
	aesKeyEnc []byte,
) error {
	body := map[string]interface{}{
		"secret_name": secretName,
		"secret_type": secretType,
		"ciphertext":  ciphertext,
		"aes_key_enc": aesKeyEnc,
	}

	resp, err := f.client.R().
		SetContext(ctx).
		SetAuthToken(secretOwner).
		SetBody(body).
		Post("/secret/save")
	if err != nil {
		return fmt.Errorf("failed to save secret: %w", err)
	}
	if resp.IsError() {
		return fmt.Errorf("secret save request returned error: %s", resp.Status())
	}

	return nil
}

// SecretReadHTTPFacade handles fetching secrets over HTTP.
type SecretReadHTTPFacade struct {
	client *resty.Client
}

func NewSecretReadHTTPFacade(client *resty.Client) *SecretReadHTTPFacade {
	return &SecretReadHTTPFacade{client: client}
}

// Get fetches a single secret by name and type and returns models.SecretDB.
func (f *SecretReadHTTPFacade) Get(
	ctx context.Context,
	secretOwner string,
	secretName string,
	secretType string,
) (*models.SecretDB, error) {
	reqBody := map[string]string{
		"secret_name": secretName,
		"secret_type": secretType,
	}

	resp, err := f.client.R().
		SetContext(ctx).
		SetAuthToken(secretOwner).
		SetBody(reqBody).
		Post("/secret/get")
	if err != nil {
		return nil, fmt.Errorf("failed to get secret: %w", err)
	}
	if resp.IsError() {
		return nil, fmt.Errorf("secret get request returned error: %s", resp.Status())
	}

	var secret models.SecretDB
	err = json.Unmarshal(resp.Body(), &secret)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal secret response: %w", err)
	}

	return &secret, nil
}

// List fetches all secrets for the authenticated user and returns a slice of pointers to models.SecretDB.
func (f *SecretReadHTTPFacade) List(
	ctx context.Context,
	secretOwner string,
) ([]*models.SecretDB, error) {
	resp, err := f.client.R().
		SetContext(ctx).
		SetAuthToken(secretOwner).
		Get("/secret/list")
	if err != nil {
		return nil, fmt.Errorf("failed to list secrets: %w", err)
	}
	if resp.IsError() {
		return nil, fmt.Errorf("secret list request returned error: %s", resp.Status())
	}

	var secrets []models.SecretDB
	err = json.Unmarshal(resp.Body(), &secrets)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal secrets list: %w", err)
	}

	// Convert slice of values to slice of pointers
	secretsPtrs := make([]*models.SecretDB, 0, len(secrets))
	for i := range secrets {
		secretsPtrs = append(secretsPtrs, &secrets[i])
	}

	return secretsPtrs, nil
}

// SecretWriteGRPCFacade handles saving secrets over gRPC.
type SecretWriteGRPCFacade struct {
	client pb.SecretWriteServiceClient
}

func NewSecretWriteGRPCFacade(conn *grpc.ClientConn) *SecretWriteGRPCFacade {
	return &SecretWriteGRPCFacade{
		client: pb.NewSecretWriteServiceClient(conn),
	}
}

// Save saves a secret with explicit parameters (no struct).
func (f *SecretWriteGRPCFacade) Save(
	ctx context.Context,
	secretOwner string,
	secretName string,
	secretType string,
	ciphertext []byte,
	aesKeyEnc []byte,
) error {
	md := metadata.Pairs("authorization", "Bearer "+secretOwner)
	ctx = metadata.NewOutgoingContext(ctx, md)

	req := &pb.SecretSaveRequest{
		SecretName: secretName,
		SecretType: secretType,
		Ciphertext: ciphertext,
		AesKeyEnc:  aesKeyEnc,
	}

	_, err := f.client.Save(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to save secret via grpc: %w", err)
	}
	return nil
}

// SecretReadGRPCFacade handles reading secrets over gRPC.
type SecretReadGRPCFacade struct {
	client pb.SecretReadServiceClient
}

func NewSecretReadGRPCFacade(conn *grpc.ClientConn) *SecretReadGRPCFacade {
	return &SecretReadGRPCFacade{
		client: pb.NewSecretReadServiceClient(conn),
	}
}

// Get fetches a single secret by name and type and returns models.SecretDB.
func (f *SecretReadGRPCFacade) Get(
	ctx context.Context,
	secretOwner string,
	secretName string,
	secretType string,
) (*models.SecretDB, error) {
	md := metadata.Pairs("authorization", "Bearer "+secretOwner)
	ctx = metadata.NewOutgoingContext(ctx, md)

	req := &pb.SecretGetRequest{
		SecretName: secretName,
		SecretType: secretType,
	}

	resp, err := f.client.Get(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get secret via grpc: %w", err)
	}

	secret := &models.SecretDB{
		SecretName:  resp.SecretName,
		SecretType:  resp.SecretType,
		SecretOwner: resp.SecretOwner,
		Ciphertext:  resp.Ciphertext,
		AESKeyEnc:   resp.AesKeyEnc,
		CreatedAt:   resp.CreatedAt.AsTime(),
		UpdatedAt:   resp.UpdatedAt.AsTime(),
	}

	return secret, nil
}

// List fetches all secrets for the authenticated user and returns a slice of pointers to models.SecretDB.
func (f *SecretReadGRPCFacade) List(
	ctx context.Context,
	secretOwner string,
) ([]*models.SecretDB, error) {
	md := metadata.Pairs("authorization", "Bearer "+secretOwner)
	ctx = metadata.NewOutgoingContext(ctx, md)

	stream, err := f.client.List(ctx, &emptypb.Empty{})
	if err != nil {
		return nil, fmt.Errorf("failed to start secret list stream: %w", err)
	}

	var secrets []*models.SecretDB
	for {
		secretResp, err := stream.Recv()
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			return nil, fmt.Errorf("error receiving secret from stream: %w", err)
		}

		secret := &models.SecretDB{
			SecretName:  secretResp.SecretName,
			SecretType:  secretResp.SecretType,
			SecretOwner: secretResp.SecretOwner,
			Ciphertext:  secretResp.Ciphertext,
			AESKeyEnc:   secretResp.AesKeyEnc,
			CreatedAt:   secretResp.CreatedAt.AsTime(),
			UpdatedAt:   secretResp.UpdatedAt.AsTime(),
		}
		secrets = append(secrets, secret)
	}

	return secrets, nil
}
