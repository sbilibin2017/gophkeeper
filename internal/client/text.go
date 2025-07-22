package client

import (
	"context"
	"errors"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/jmoiron/sqlx"
	"github.com/sbilibin2017/gophkeeper/internal/models"
	"github.com/sbilibin2017/gophkeeper/internal/models/fields"
	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

// TextAddClient inserts or updates a text secret in the local database.
func TextAddClient(
	ctx context.Context,
	db *sqlx.DB,
	req *models.TextAddRequest,
) error {
	query := `
		INSERT INTO text_client (secret_name, content, meta, updated_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (secret_name) DO UPDATE SET
			content = EXCLUDED.content,
			meta = EXCLUDED.meta,
			updated_at = EXCLUDED.updated_at
	`

	_, err := db.ExecContext(ctx, query,
		req.SecretName,
		req.Content,
		req.Meta, // *fields.StringMap implements driver.Valuer
		time.Now().UTC(),
	)

	return err
}

// TextGetClient retrieves a text secret from the local database by secret name.
func TextGetClient(
	ctx context.Context,
	db *sqlx.DB,
	secretName string,
) (*models.TextDB, error) {
	query := `
		SELECT secret_name, content, meta, updated_at
		FROM text_client
		WHERE secret_name = $1
		LIMIT 1
	`

	var text models.TextDB
	err := db.GetContext(ctx, &text, query, secretName)
	if err != nil {
		return nil, err
	}
	return &text, nil
}

// TextListClient retrieves all text secrets from the local database.
func TextListClient(
	ctx context.Context,
	db *sqlx.DB,
) ([]models.TextDB, error) {
	query := `
		SELECT secret_name, content, meta, updated_at
		FROM text_client
	`

	var texts []models.TextDB
	err := db.SelectContext(ctx, &texts, query)
	if err != nil {
		return nil, err
	}

	return texts, nil
}

// TextGetHTTP retrieves a text secret from a remote HTTP server by secret name.
func TextGetHTTP(
	ctx context.Context,
	client *resty.Client,
	secretName string,
) (*models.TextDB, error) {
	resp := &models.TextDB{}
	r, err := client.R().
		SetContext(ctx).
		SetResult(resp).
		Get("/text/" + secretName)
	if err != nil {
		return nil, err
	}
	if r.IsError() {
		return nil, errors.New(r.Status())
	}
	return resp, nil
}

// TextGetGRPC retrieves a text secret from a remote gRPC server by secret name.
func TextGetGRPC(
	ctx context.Context,
	client pb.TextServiceClient,
	secretName string,
) (*models.TextDB, error) {
	grpcReq := &pb.TextFilterRequest{
		SecretName: secretName,
	}

	grpcResp, err := client.Get(ctx, grpcReq)
	if err != nil {
		return nil, err
	}

	var updatedAt time.Time
	if grpcResp.UpdatedAt != nil {
		updatedAt = grpcResp.UpdatedAt.AsTime()
	}

	var meta *fields.StringMap
	if grpcResp.Meta != nil {
		sm := &fields.StringMap{
			Map: make(map[string]string, len(grpcResp.Meta)),
		}
		for k, v := range grpcResp.Meta {
			sm.Map[k] = v
		}
		meta = sm
	}

	return &models.TextDB{
		SecretName: grpcResp.SecretName,
		Content:    grpcResp.Content,
		Meta:       meta,
		UpdatedAt:  updatedAt,
	}, nil
}

// TextAddGRPC sends a request to a gRPC server to add or update a text secret.
func TextAddGRPC(
	ctx context.Context,
	client pb.TextServiceClient,
	req *models.TextAddRequest,
) error {
	grpcReq := &pb.TextAddRequest{
		SecretName: req.SecretName,
		Content:    req.Content,
		Meta:       make(map[string]string),
	}

	if req.Meta != nil && req.Meta.Map != nil {
		for k, v := range req.Meta.Map {
			grpcReq.Meta[k] = v
		}
	}

	_, err := client.Add(ctx, grpcReq)
	return err
}

// TextAddHTTP sends a POST request to a remote HTTP server to add or update a text secret.
func TextAddHTTP(
	ctx context.Context,
	client *resty.Client,
	req *models.TextAddRequest,
) error {
	r, err := client.R().
		SetContext(ctx).
		SetBody(req).
		Post("/text")
	if err != nil {
		return err
	}
	if r.IsError() {
		return errors.New(r.Status())
	}
	return nil
}

// TextListHTTP fetches a list of all text secrets from the remote HTTP API.
func TextListHTTP(
	ctx context.Context,
	client *resty.Client,
) ([]models.TextDB, error) {
	var resp []models.TextDB
	r, err := client.R().
		SetContext(ctx).
		SetResult(&resp).
		Get("/text/")
	if err != nil {
		return nil, err
	}
	if r.IsError() {
		return nil, errors.New(r.Status())
	}
	return resp, nil
}

// TextListGRPC fetches a list of all text secrets from the remote gRPC service using stream.
func TextListGRPC(
	ctx context.Context,
	client pb.TextServiceClient,
) ([]models.TextDB, error) {
	stream, err := client.List(ctx, &emptypb.Empty{})
	if err != nil {
		return nil, err
	}

	var texts []models.TextDB

	for {
		grpcText, err := stream.Recv()
		if err != nil {
			if err.Error() == "EOF" {
				break // stream complete
			}
			return nil, err
		}

		var updatedAt time.Time
		if grpcText.UpdatedAt != nil {
			updatedAt = grpcText.UpdatedAt.AsTime()
		}

		var meta *fields.StringMap
		if grpcText.Meta != nil {
			sm := &fields.StringMap{
				Map: make(map[string]string, len(grpcText.Meta)),
			}
			for k, v := range grpcText.Meta {
				sm.Map[k] = v
			}
			meta = sm
		}

		text := models.TextDB{
			SecretName: grpcText.SecretName,
			Content:    grpcText.Content,
			Meta:       meta,
			UpdatedAt:  updatedAt,
		}

		texts = append(texts, text)
	}

	return texts, nil
}
