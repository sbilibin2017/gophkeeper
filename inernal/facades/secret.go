package facades

import (
	"context"
	"fmt"
	"io"

	"github.com/go-resty/resty/v2"
	"google.golang.org/grpc"

	"github.com/sbilibin2017/gophkeeper/inernal/models"
	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc"
)

// SecretHTTPWriteFacade is an HTTP client facade for writing (saving) secrets via REST API.
type SecretHTTPWriteFacade struct {
	client *resty.Client
}

// NewSecretHTTPWriteFacade creates a new SecretHTTPWriteFacade with the given Resty client.
func NewSecretHTTPWriteFacade(client *resty.Client) *SecretHTTPWriteFacade {
	return &SecretHTTPWriteFacade{client: client}
}

// Save sends a secret save request over HTTP.
func (f *SecretHTTPWriteFacade) Save(ctx context.Context, secret *models.SecretSaveRequest) error {
	resp, err := f.client.R().
		SetContext(ctx).
		SetAuthToken(secret.Token).
		SetBody(secret).
		Post("/save/")
	if err != nil {
		return fmt.Errorf("http save request failed: %w", err)
	}
	if resp.IsError() {
		return fmt.Errorf("http error status %d, body: %s", resp.StatusCode(), resp.String())
	}
	return nil
}

// SecretHTTPReadFacade is an HTTP client facade for reading secrets via REST API.
type SecretHTTPReadFacade struct {
	client *resty.Client
}

// NewSecretHTTPReadFacade creates a new SecretHTTPReadFacade with the given Resty client.
func NewSecretHTTPReadFacade(client *resty.Client) *SecretHTTPReadFacade {
	return &SecretHTTPReadFacade{client: client}
}

// Get retrieves a secret over HTTP using secret type and name.
func (f *SecretHTTPReadFacade) Get(ctx context.Context, req *models.SecretGetRequest) (*models.SecretDB, error) {
	var secret models.SecretDB

	resp, err := f.client.R().
		SetContext(ctx).
		SetAuthToken(req.Token).
		SetResult(&secret).
		SetPathParam("secretType", req.SecretType).
		SetPathParam("secretName", req.SecretName).
		Get("/get/{secretType}/{secretName}")
	if err != nil {
		return nil, fmt.Errorf("http get request failed: %w", err)
	}
	if resp.IsError() {
		return nil, fmt.Errorf("http error status %d, body: %s", resp.StatusCode(), resp.String())
	}
	return &secret, nil
}

// List retrieves all secrets accessible by the token via HTTP.
func (f *SecretHTTPReadFacade) List(ctx context.Context, req *models.SecretListRequest) ([]*models.SecretDB, error) {
	var secrets []*models.SecretDB

	resp, err := f.client.R().
		SetContext(ctx).
		SetAuthToken(req.Token).
		SetResult(&secrets).
		Get("/list/")
	if err != nil {
		return nil, fmt.Errorf("http list request failed: %w", err)
	}
	if resp.IsError() {
		return nil, fmt.Errorf("http error status %d, body: %s", resp.StatusCode(), resp.String())
	}
	return secrets, nil
}

// SecretGRPCWriteFacade is a gRPC client facade for writing (saving) secrets.
type SecretGRPCWriteFacade struct {
	client pb.SecretWriteServiceClient
}

// NewSecretGRPCWriteFacade creates a new SecretGRPCWriteFacade from a gRPC ClientConn.
// It internally creates the SecretWriteServiceClient.
func NewSecretGRPCWriteFacade(conn *grpc.ClientConn) *SecretGRPCWriteFacade {
	client := pb.NewSecretWriteServiceClient(conn)
	return &SecretGRPCWriteFacade{client: client}
}

// Save sends a secret save request via gRPC.
func (f *SecretGRPCWriteFacade) Save(ctx context.Context, secret *models.SecretSaveRequest) error {
	req := &pb.SecretSaveRequest{
		SecretName: secret.SecretName,
		SecretType: secret.SecretType,
		Ciphertext: secret.Ciphertext,
		AesKeyEnc:  secret.AESKeyEnc,
		Token:      secret.Token,
	}

	_, err := f.client.Save(ctx, req)
	if err != nil {
		return fmt.Errorf("grpc save failed: %w", err)
	}
	return nil
}

// SecretGRPCReadFacade is a gRPC client facade for reading secrets.
type SecretGRPCReadFacade struct {
	client pb.SecretReadServiceClient
}

// NewSecretGRPCReadFacade creates a new SecretGRPCReadFacade from a gRPC ClientConn.
// It internally creates the SecretReadServiceClient.
func NewSecretGRPCReadFacade(conn *grpc.ClientConn) *SecretGRPCReadFacade {
	client := pb.NewSecretReadServiceClient(conn)
	return &SecretGRPCReadFacade{client: client}
}

// Get retrieves a secret via gRPC using secret type and name.
func (f *SecretGRPCReadFacade) Get(ctx context.Context, req *models.SecretGetRequest) (*models.SecretDB, error) {
	pbReq := &pb.SecretGetRequest{
		SecretName: req.SecretName,
		SecretType: req.SecretType,
		Token:      req.Token,
	}

	pbResp, err := f.client.Get(ctx, pbReq)
	if err != nil {
		return nil, fmt.Errorf("grpc get secret failed: %w", err)
	}

	secret := &models.SecretDB{
		SecretName: pbResp.SecretName,
		SecretType: pbResp.SecretType,
		Ciphertext: pbResp.Ciphertext,
		AESKeyEnc:  pbResp.AesKeyEnc,
		CreatedAt:  pbResp.CreatedAt.AsTime(),
		UpdatedAt:  pbResp.UpdatedAt.AsTime(),
	}

	return secret, nil
}

// List retrieves all secrets accessible by the token via a gRPC streaming call.
func (f *SecretGRPCReadFacade) List(ctx context.Context, req *models.SecretListRequest) ([]*models.SecretDB, error) {
	stream, err := f.client.List(ctx, &pb.SecretListRequest{Token: req.Token})
	if err != nil {
		return nil, fmt.Errorf("grpc list secrets failed to start stream: %w", err)
	}

	var secrets []*models.SecretDB
	for {
		pbResp, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("grpc list secrets stream error: %w", err)
		}

		secret := &models.SecretDB{
			SecretName: pbResp.SecretName,
			SecretType: pbResp.SecretType,
			Ciphertext: pbResp.Ciphertext,
			AESKeyEnc:  pbResp.AesKeyEnc,
			CreatedAt:  pbResp.CreatedAt.AsTime(),
			UpdatedAt:  pbResp.UpdatedAt.AsTime(),
		}
		secrets = append(secrets, secret)
	}

	return secrets, nil
}
