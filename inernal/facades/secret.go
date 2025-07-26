package facades

import (
	"context"
	"fmt"
	"io"

	"github.com/go-resty/resty/v2"
	"github.com/sbilibin2017/gophkeeper/inernal/models"
	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc"
	"google.golang.org/grpc"
)

// SecretHTTPWriteFacade provides HTTP client methods to write secrets.
type SecretHTTPWriteFacade struct {
	client *resty.Client
}

// NewSecretHTTPWriteFacade creates a new SecretHTTPWriteFacade with the given Resty client.
func NewSecretHTTPWriteFacade(client *resty.Client) *SecretHTTPWriteFacade {
	return &SecretHTTPWriteFacade{client: client}
}

// Save sends a secret save request via HTTP using the provided context and secret data.
func (f *SecretHTTPWriteFacade) Save(ctx context.Context, secret *models.SecretSaveRequest) error {
	resp, err := f.client.R().
		SetContext(ctx).
		SetAuthToken(secret.Token).
		SetBody(secret).
		Post("/save/")
	if err != nil {
		return err
	}
	if resp.IsError() {
		return fmt.Errorf("status %d, body: %s", resp.StatusCode(), resp.String())
	}
	return nil
}

// SecretHTTPReadFacade provides HTTP client methods to read secrets.
type SecretHTTPReadFacade struct {
	client *resty.Client
}

// NewSecretHTTPReadFacade creates a new SecretHTTPReadFacade with the given Resty client.
func NewSecretHTTPReadFacade(client *resty.Client) *SecretHTTPReadFacade {
	return &SecretHTTPReadFacade{client: client}
}

// Get retrieves a secret by its type and name over HTTP using the given context and request parameters.
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
		return nil, err
	}
	if resp.IsError() {
		return nil, fmt.Errorf("status %d, body: %s", resp.StatusCode(), resp.String())
	}
	return &secret, nil
}

// List fetches all secrets accessible by the provided token over HTTP.
func (f *SecretHTTPReadFacade) List(ctx context.Context, req *models.SecretListRequest) ([]*models.SecretDB, error) {
	var secrets []*models.SecretDB

	resp, err := f.client.R().
		SetContext(ctx).
		SetHeader("Authorization", "Bearer "+req.Token).
		SetResult(&secrets).
		Get("/list/")
	if err != nil {
		return nil, err
	}
	if resp.IsError() {
		return nil, fmt.Errorf("status %d, body: %s", resp.StatusCode(), resp.String())
	}
	return secrets, nil
}

// SecretGRPCWriteFacade provides gRPC client methods to write secrets.
type SecretGRPCWriteFacade struct {
	client pb.SecretWriteServiceClient
}

// NewSecretGRPCWriteFacade creates a new SecretGRPCWriteFacade using the provided gRPC client connection.
func NewSecretGRPCWriteFacade(conn *grpc.ClientConn) *SecretGRPCWriteFacade {
	return &SecretGRPCWriteFacade{
		client: pb.NewSecretWriteServiceClient(conn),
	}
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
	return err
}

// SecretGRPCReadFacade provides gRPC client methods to read secrets.
type SecretGRPCReadFacade struct {
	client pb.SecretReadServiceClient
}

// NewSecretGRPCReadFacade creates a new SecretGRPCReadFacade using the provided gRPC client connection.
func NewSecretGRPCReadFacade(conn *grpc.ClientConn) *SecretGRPCReadFacade {
	return &SecretGRPCReadFacade{
		client: pb.NewSecretReadServiceClient(conn),
	}
}

// Get retrieves a secret by type and name via gRPC.
func (f *SecretGRPCReadFacade) Get(ctx context.Context, req *models.SecretGetRequest) (*models.SecretDB, error) {
	grpcReq := &pb.SecretGetRequest{
		SecretName: req.SecretName,
		SecretType: req.SecretType,
		Token:      req.Token,
	}

	resp, err := f.client.Get(ctx, grpcReq)
	if err != nil {
		return nil, err
	}

	return &models.SecretDB{
		SecretName:  resp.SecretName,
		SecretType:  resp.SecretType,
		SecretOwner: resp.SecretOwner,
		Ciphertext:  resp.Ciphertext,
		AESKeyEnc:   resp.AesKeyEnc,
		CreatedAt:   resp.CreatedAt.AsTime(),
		UpdatedAt:   resp.UpdatedAt.AsTime(),
	}, nil
}

// List streams all secrets accessible by the token via gRPC.
func (f *SecretGRPCReadFacade) List(ctx context.Context, req *models.SecretListRequest) ([]*models.SecretDB, error) {
	stream, err := f.client.List(ctx, &pb.SecretListRequest{Token: req.Token})
	if err != nil {
		return nil, err
	}

	var secrets []*models.SecretDB
	for {
		secretProto, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		secrets = append(secrets, &models.SecretDB{
			SecretName:  secretProto.SecretName,
			SecretType:  secretProto.SecretType,
			SecretOwner: secretProto.SecretOwner,
			Ciphertext:  secretProto.Ciphertext,
			AESKeyEnc:   secretProto.AesKeyEnc,
			CreatedAt:   secretProto.CreatedAt.AsTime(),
			UpdatedAt:   secretProto.UpdatedAt.AsTime(),
		})
	}

	return secrets, nil
}
