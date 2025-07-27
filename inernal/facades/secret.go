package facades

import (
	"context"
	"fmt"
	"io"

	"github.com/go-resty/resty/v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/sbilibin2017/gophkeeper/inernal/models"
	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc"
)

type SecretWriteHTTP struct {
	client *resty.Client
}

func NewSecretWriteHTTP(client *resty.Client) *SecretWriteHTTP {
	return &SecretWriteHTTP{client: client}
}

func (w *SecretWriteHTTP) Save(
	ctx context.Context,
	token string,
	secret *models.Secret,
) error {
	resp, err := w.client.R().
		SetContext(ctx).
		SetAuthToken(token).
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

type SecretReadHTTP struct {
	client *resty.Client
}

func NewSecretReadHTTP(client *resty.Client) *SecretReadHTTP {
	return &SecretReadHTTP{client: client}
}

func (r *SecretReadHTTP) Get(
	ctx context.Context,
	token string,
	secretName string,
	secretType string,
) (*models.Secret, error) {
	var secret models.Secret

	resp, err := r.client.R().
		SetContext(ctx).
		SetAuthToken(token).
		SetResult(&secret).
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

func (r *SecretReadHTTP) List(
	ctx context.Context,
	token string,
) ([]*models.Secret, error) {
	var secrets []*models.Secret

	resp, err := r.client.R().
		SetContext(ctx).
		SetAuthToken(token).
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

type SecretWriteGRPC struct {
	client pb.SecretWriteServiceClient
}

func NewSecretWriteGRPC(conn *grpc.ClientConn) *SecretWriteGRPC {
	client := pb.NewSecretWriteServiceClient(conn)
	return &SecretWriteGRPC{client: client}
}

func (w *SecretWriteGRPC) Save(
	ctx context.Context,
	token string,
	secret *models.Secret,
) error {
	ctx = metadata.AppendToOutgoingContext(ctx, "authorization", fmt.Sprintf("Bearer %s", token))

	pbReq := &pb.Secret{
		SecretName:  secret.SecretName,
		SecretType:  secret.SecretType,
		SecretOwner: secret.SecretOwner,
		Ciphertext:  secret.Ciphertext,
		AesKeyEnc:   secret.AESKeyEnc,
		CreatedAt:   timestamppb.New(secret.CreatedAt),
		UpdatedAt:   timestamppb.New(secret.UpdatedAt),
	}

	_, err := w.client.Save(ctx, pbReq)
	if err != nil {
		return fmt.Errorf("grpc save failed: %w", err)
	}
	return nil
}

type SecretReadGRPC struct {
	client pb.SecretReadServiceClient
}

func NewSecretReadGRPC(conn *grpc.ClientConn) *SecretReadGRPC {
	client := pb.NewSecretReadServiceClient(conn)
	return &SecretReadGRPC{client: client}
}

func (r *SecretReadGRPC) Get(
	ctx context.Context,
	token string,
	secretName string,
	secretType string,
) (*models.Secret, error) {
	ctx = metadata.AppendToOutgoingContext(
		ctx,
		"authorization", fmt.Sprintf("Bearer %s", token),
	)

	pbReq := &pb.SecretGetRequest{
		SecretName: secretName,
		SecretType: secretType,
	}

	pbResp, err := r.client.Get(ctx, pbReq)
	if err != nil {
		return nil, fmt.Errorf("grpc get secret failed: %w", err)
	}

	return &models.Secret{
		SecretName:  pbResp.SecretName,
		SecretType:  pbResp.SecretType,
		SecretOwner: pbResp.SecretOwner,
		Ciphertext:  pbResp.Ciphertext,
		AESKeyEnc:   pbResp.AesKeyEnc,
		CreatedAt:   pbResp.CreatedAt.AsTime(),
		UpdatedAt:   pbResp.UpdatedAt.AsTime(),
	}, nil
}

func (r *SecretReadGRPC) List(
	ctx context.Context,
	token string,
) ([]*models.Secret, error) {
	ctx = metadata.AppendToOutgoingContext(ctx, "authorization", fmt.Sprintf("Bearer %s", token))

	stream, err := r.client.List(ctx, &emptypb.Empty{})
	if err != nil {
		return nil, fmt.Errorf("grpc list secrets failed to start stream: %w", err)
	}

	var secrets []*models.Secret
	for {
		pbResp, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("grpc list secrets stream error: %w", err)
		}

		secrets = append(secrets, &models.Secret{
			SecretName:  pbResp.SecretName,
			SecretType:  pbResp.SecretType,
			SecretOwner: pbResp.SecretOwner,
			Ciphertext:  pbResp.Ciphertext,
			AESKeyEnc:   pbResp.AesKeyEnc,
			CreatedAt:   pbResp.CreatedAt.AsTime(),
			UpdatedAt:   pbResp.UpdatedAt.AsTime(),
		})
	}
	return secrets, nil
}
