package facades

import (
	"context"
	"fmt"

	"github.com/go-resty/resty/v2"
	"github.com/sbilibin2017/gophkeeper/internal/models"
	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

// SecretHTTPWriteFacade provides an HTTP client facade for writing secrets.
type SecretHTTPWriteFacade struct {
	client *resty.Client
}

// NewSecretHTTPWriteFacade creates a new SecretHTTPWriteFacade with a resty.Client.
func NewSecretHTTPWriteFacade(client *resty.Client) *SecretHTTPWriteFacade {
	return &SecretHTTPWriteFacade{
		client: client,
	}
}

// Save sends a new encrypted secret to the server via HTTP POST.
func (f *SecretHTTPWriteFacade) Save(ctx context.Context, secret *models.EncryptedSecret) error {
	resp, err := f.client.R().
		SetContext(ctx).
		SetBody(secret).
		Post("/secret")
	if err != nil {
		return err
	}
	if resp.IsError() {
		return fmt.Errorf("failed to add secret: status %d, body: %s", resp.StatusCode(), resp.String())
	}
	return nil
}

// SecretHTTPReadFacade provides an HTTP client facade for reading secrets.
type SecretHTTPReadFacade struct {
	client *resty.Client
}

// NewSecretHTTPReadFacade creates a new SecretHTTPReadFacade with a resty.Client.
func NewSecretHTTPReadFacade(client *resty.Client) *SecretHTTPReadFacade {
	return &SecretHTTPReadFacade{
		client: client,
	}
}

// Get retrieves a single encrypted secret by its secret name via HTTP GET.
func (f *SecretHTTPReadFacade) Get(ctx context.Context, secretName string) (*models.EncryptedSecret, error) {
	var secret models.EncryptedSecret

	resp, err := f.client.R().
		SetContext(ctx).
		SetResult(&secret).
		SetPathParam("secretName", secretName).
		Get("/secret/{secretName}")
	if err != nil {
		return nil, err
	}

	if resp.IsError() {
		return nil, fmt.Errorf("failed to get secret '%s': status %d, body: %s", secretName, resp.StatusCode(), resp.String())
	}

	return &secret, nil
}

// List fetches all encrypted secrets from the server via HTTP GET.
func (f *SecretHTTPReadFacade) List(ctx context.Context) ([]*models.EncryptedSecret, error) {
	var secrets []*models.EncryptedSecret

	resp, err := f.client.R().
		SetContext(ctx).
		SetResult(&secrets).
		Get("/secrets/")
	if err != nil {
		return nil, err
	}

	if resp.IsError() {
		return nil, fmt.Errorf("failed to list secrets: status %d, body: %s", resp.StatusCode(), resp.String())
	}

	return secrets, nil
}

// SecretGRPCWriteFacade provides a gRPC client facade for writing secrets.
type SecretGRPCWriteFacade struct {
	client pb.SecretWriteServiceClient
}

// NewSecretGRPCWriteFacade creates a new SecretGRPCWriteFacade using a gRPC connection.
func NewSecretGRPCWriteFacade(conn *grpc.ClientConn) *SecretGRPCWriteFacade {
	return &SecretGRPCWriteFacade{
		client: pb.NewSecretWriteServiceClient(conn),
	}
}

// Add sends a new encrypted secret to the gRPC service.
func (f *SecretGRPCWriteFacade) Save(ctx context.Context, secret *models.EncryptedSecret) error {
	req := &pb.EncryptedSecret{
		SecretName: secret.SecretName,
		SecretType: secret.SecretType,
		Ciphertext: secret.Ciphertext,
		AesKeyEnc:  secret.AESKeyEnc,
		Timestamp:  secret.Timestamp,
	}

	_, err := f.client.Save(ctx, req)
	return err
}

// SecretGRPCReadFacade provides a gRPC client facade for reading secrets.
type SecretGRPCReadFacade struct {
	client pb.SecretReadServiceClient
}

// NewSecretGRPCReadFacade creates a new SecretGRPCReadFacade using a gRPC connection.
func NewSecretGRPCReadFacade(conn *grpc.ClientConn) *SecretGRPCReadFacade {
	return &SecretGRPCReadFacade{
		client: pb.NewSecretReadServiceClient(conn),
	}
}

// Get retrieves a single encrypted secret by its name from the gRPC service.
func (f *SecretGRPCReadFacade) Get(ctx context.Context, secretName string) (*models.EncryptedSecret, error) {
	req := &pb.GetSecretRequest{
		SecretName: secretName,
	}

	resp, err := f.client.Get(ctx, req)
	if err != nil {
		return nil, err
	}

	return &models.EncryptedSecret{
		SecretName: resp.SecretName,
		SecretType: resp.SecretType,
		Ciphertext: resp.Ciphertext,
		AESKeyEnc:  resp.AesKeyEnc,
		Timestamp:  resp.Timestamp,
	}, nil
}

// List streams all encrypted secrets from the gRPC service.
func (f *SecretGRPCReadFacade) List(ctx context.Context) ([]*models.EncryptedSecret, error) {
	stream, err := f.client.List(ctx, &emptypb.Empty{})
	if err != nil {
		return nil, err
	}

	var secrets []*models.EncryptedSecret
	for {
		secretProto, err := stream.Recv()
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			return nil, err
		}

		secrets = append(secrets, &models.EncryptedSecret{
			SecretName: secretProto.SecretName,
			SecretType: secretProto.SecretType,
			Ciphertext: secretProto.Ciphertext,
			AESKeyEnc:  secretProto.AesKeyEnc,
			Timestamp:  secretProto.Timestamp,
		})
	}

	return secrets, nil
}
