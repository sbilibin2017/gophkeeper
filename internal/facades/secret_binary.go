package facades

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/sbilibin2017/gophkeeper/internal/models"
	"google.golang.org/grpc/metadata"

	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc"
)

// --- HTTP фасады ---

type SecretBinaryListFacade struct {
	client *resty.Client
}

func NewSecretBinaryListFacade(client *resty.Client) *SecretBinaryListFacade {
	return &SecretBinaryListFacade{client: client}
}

func (f *SecretBinaryListFacade) List(ctx context.Context, token string) ([]models.SecretBinaryClient, error) {
	var secrets []models.SecretBinaryClient

	resp, err := f.client.R().
		SetContext(ctx).
		SetAuthToken(token).
		SetResult(&secrets).
		Get("/list/secret-binary")
	if err != nil {
		return nil, fmt.Errorf("server unavailable: %w", err)
	}
	if resp.IsError() {
		return nil, errors.New("error fetching binary secrets: " + resp.Status())
	}
	return secrets, nil
}

type SecretBinarySaveHTTPFacade struct {
	client *resty.Client
}

func NewSecretBinarySaveHTTPFacade(client *resty.Client) *SecretBinarySaveHTTPFacade {
	return &SecretBinarySaveHTTPFacade{client: client}
}

func (f *SecretBinarySaveHTTPFacade) Save(ctx context.Context, token string, secret models.SecretBinaryClient) error {
	reqBody := map[string]interface{}{
		"secret_name": secret.SecretName,
		"data":        secret.Data,
		"updated_at":  secret.UpdatedAt.Format(time.RFC3339),
	}

	if secret.Meta != nil {
		reqBody["meta"] = *secret.Meta
	}

	resp, err := f.client.R().
		SetContext(ctx).
		SetAuthToken(token).
		SetBody(reqBody).
		Post("/save/secret-binary")

	if err != nil {
		return fmt.Errorf("failed to send save request: %w", err)
	}

	if resp.IsError() {
		return errors.New("server error: " + resp.Status())
	}

	return nil
}

type SecretBinaryGetHTTPFacade struct {
	client *resty.Client
}

func NewSecretBinaryGetHTTPFacade(client *resty.Client) *SecretBinaryGetHTTPFacade {
	return &SecretBinaryGetHTTPFacade{client: client}
}

func (f *SecretBinaryGetHTTPFacade) Get(ctx context.Context, token string, secretName string) (*models.SecretBinaryClient, error) {
	var secret models.SecretBinaryClient

	url := fmt.Sprintf("/get/secret-binary/%s", secretName)

	resp, err := f.client.R().
		SetContext(ctx).
		SetAuthToken(token).
		SetResult(&secret).
		Get(url)
	if err != nil {
		return nil, fmt.Errorf("server unavailable: %w", err)
	}
	if resp.IsError() {
		return nil, errors.New("error fetching binary secret: " + resp.Status())
	}

	return &secret, nil
}

// --- gRPC фасады ---

type SecretBinaryListGRPCFacade struct {
	client pb.SecretBinaryServiceClient
}

func NewSecretBinaryListGRPCFacade(client pb.SecretBinaryServiceClient) *SecretBinaryListGRPCFacade {
	return &SecretBinaryListGRPCFacade{client: client}
}

func (f *SecretBinaryListGRPCFacade) List(ctx context.Context, token string) ([]models.SecretBinaryClient, error) {
	md := metadata.New(map[string]string{
		"authorization": "Bearer " + token,
	})
	ctx = metadata.NewOutgoingContext(ctx, md)

	resp, err := f.client.List(ctx, &pb.SecretBinaryListRequest{})
	if err != nil {
		return nil, err
	}

	var result []models.SecretBinaryClient
	for _, item := range resp.Items {
		updatedAt, err := time.Parse(time.RFC3339, item.UpdatedAt)
		if err != nil {
			return nil, errors.New("invalid updated_at format in response")
		}

		var meta *string
		if item.Meta != "" {
			meta = &item.Meta
		}

		result = append(result, models.SecretBinaryClient{
			SecretName: item.SecretName,
			Data:       item.Data,
			Meta:       meta,
			UpdatedAt:  updatedAt,
		})
	}

	return result, nil
}

type SecretBinaryGetGRPCFacade struct {
	client pb.SecretBinaryServiceClient
}

func NewSecretBinaryGetGRPCFacade(client pb.SecretBinaryServiceClient) *SecretBinaryGetGRPCFacade {
	return &SecretBinaryGetGRPCFacade{client: client}
}

func (f *SecretBinaryGetGRPCFacade) Get(ctx context.Context, token string, secretName string) (*models.SecretBinaryClient, error) {
	md := metadata.New(map[string]string{
		"authorization": "Bearer " + token,
	})
	ctx = metadata.NewOutgoingContext(ctx, md)

	req := &pb.SecretBinaryGetRequest{
		SecretName: secretName, // убрал Token из тела
	}

	resp, err := f.client.Get(ctx, req)
	if err != nil {
		return nil, err
	}

	item := resp.Secret
	updatedAt, err := time.Parse(time.RFC3339, item.UpdatedAt)
	if err != nil {
		return nil, errors.New("invalid updated_at format in response")
	}

	var meta *string
	if item.Meta != "" {
		meta = &item.Meta
	}

	return &models.SecretBinaryClient{
		SecretName: item.SecretName,
		Data:       item.Data,
		Meta:       meta,
		UpdatedAt:  updatedAt,
	}, nil
}

type SecretBinarySaveGRPCFacade struct {
	client pb.SecretBinaryServiceClient
}

func NewSecretBinarySaveGRPCFacade(client pb.SecretBinaryServiceClient) *SecretBinarySaveGRPCFacade {
	return &SecretBinarySaveGRPCFacade{client: client}
}

func (f *SecretBinarySaveGRPCFacade) Save(ctx context.Context, token string, secret models.SecretBinaryClient) error {
	md := metadata.New(map[string]string{
		"authorization": "Bearer " + token,
	})
	ctx = metadata.NewOutgoingContext(ctx, md)

	meta := ""
	if secret.Meta != nil {
		meta = *secret.Meta
	}

	req := &pb.SecretBinarySaveRequest{
		Secret: &pb.SecretBinary{
			SecretName: secret.SecretName,
			Data:       secret.Data,
			Meta:       meta,
			UpdatedAt:  secret.UpdatedAt.Format(time.RFC3339),
		},
	}

	_, err := f.client.Save(ctx, req)
	return err
}
