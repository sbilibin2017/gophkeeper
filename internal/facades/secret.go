package facades

import (
	"context"
	"fmt"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/sbilibin2017/gophkeeper/internal/models"
	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// SecretHTTPWriteFacade provides an HTTP client facade for writing secrets.
type SecretHTTPWriteFacade struct {
	client *resty.Client
}

// Add sends a new encrypted secret to the server via HTTP POST.
func (f *SecretHTTPWriteFacade) Add(
	ctx context.Context, secret *models.EncrypedSecret,
) error {
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

// Get retrieves a single encrypted secret by its secret name via HTTP GET.
func (f *SecretHTTPReadFacade) Get(
	ctx context.Context, secretName string,
) (*models.EncrypedSecret, error) {
	var secret models.EncrypedSecret

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
func (f *SecretHTTPReadFacade) List(
	ctx context.Context,
) ([]*models.EncrypedSecret, error) {
	var secrets []*models.EncrypedSecret

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
func (f *SecretGRPCWriteFacade) Add(ctx context.Context, secret *models.EncrypedSecret) error {
	req := &pb.EncrypedSecret{
		SecretName: secret.SecretName,
		SecretType: secret.SecretType,
		Ciphertext: secret.Ciphertext,
		Hmac:       secret.HMAC,
		Nonce:      secret.Nonce,
		KeyId:      secret.KeyID,
	}
	if secret.Timestamp != 0 {
		req.Timestamp = timestamppb.New(time.Unix(secret.Timestamp, 0))
	}

	_, err := f.client.Add(ctx, req)
	if err != nil {
		return err
	}
	return nil
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
func (f *SecretGRPCReadFacade) Get(ctx context.Context, secretName string) (*models.EncrypedSecret, error) {
	req := &pb.GetSecretRequest{
		SecretName: secretName,
	}

	resp, err := f.client.Get(ctx, req)
	if err != nil {
		return nil, err
	}

	var ts int64
	if resp.Timestamp != nil {
		ts = resp.Timestamp.Seconds
	}

	return &models.EncrypedSecret{
		SecretName: resp.SecretName,
		SecretType: resp.SecretType,
		Ciphertext: resp.Ciphertext,
		HMAC:       resp.Hmac,
		Nonce:      resp.Nonce,
		KeyID:      resp.KeyId,
		Timestamp:  ts,
	}, nil
}

// List streams all encrypted secrets from the gRPC service.
func (f *SecretGRPCReadFacade) List(ctx context.Context) ([]*models.EncrypedSecret, error) {
	stream, err := f.client.List(ctx, &emptypb.Empty{})
	if err != nil {
		return nil, err
	}

	var secrets []*models.EncrypedSecret
	for {
		secretProto, err := stream.Recv()
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			return nil, err
		}

		var ts int64
		if secretProto.Timestamp != nil {
			ts = secretProto.Timestamp.Seconds
		}

		secrets = append(secrets, &models.EncrypedSecret{
			SecretName: secretProto.SecretName,
			SecretType: secretProto.SecretType,
			Ciphertext: secretProto.Ciphertext,
			HMAC:       secretProto.Hmac,
			Nonce:      secretProto.Nonce,
			KeyID:      secretProto.KeyId,
			Timestamp:  ts,
		})
	}

	return secrets, nil
}
