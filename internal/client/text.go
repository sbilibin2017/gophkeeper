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
)

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
