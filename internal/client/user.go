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

func UserAddClient(
	ctx context.Context,
	db *sqlx.DB,
	req *models.UserAddRequest,
) error {
	query := `
		INSERT INTO user_client (secret_name, username, password, meta, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (secret_name) DO UPDATE SET
			username = EXCLUDED.username,
			password = EXCLUDED.password,
			meta = EXCLUDED.meta,
			updated_at = EXCLUDED.updated_at
	`

	_, err := db.ExecContext(ctx, query,
		req.SecretName,
		req.Username,
		req.Password,
		req.Meta, // *fields.StringMap implements driver.Valuer
		time.Now().UTC(),
	)

	return err
}

func UserGetClient(
	ctx context.Context,
	db *sqlx.DB,
	secretName string,
) (*models.UserDB, error) {
	query := `
		SELECT secret_name, username, password, meta, updated_at
		FROM user_client
		WHERE secret_name = $1
		LIMIT 1
	`

	var user models.UserDB
	err := db.GetContext(ctx, &user, query, secretName)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func UserListClient(
	ctx context.Context,
	db *sqlx.DB,
) ([]models.UserDB, error) {
	query := `
		SELECT secret_name, username, password, meta, updated_at
		FROM user_client
	`

	var users []models.UserDB
	err := db.SelectContext(ctx, &users, query)
	if err != nil {
		return nil, err
	}

	return users, nil
}

func UserGetHTTP(
	ctx context.Context,
	client *resty.Client,
	secretName string,
) (*models.UserDB, error) {
	resp := &models.UserDB{}
	r, err := client.R().
		SetContext(ctx).
		SetResult(resp).
		Get("/user/" + secretName)
	if err != nil {
		return nil, err
	}
	if r.IsError() {
		return nil, errors.New(r.Status())
	}
	return resp, nil
}

func UserGetGRPC(
	ctx context.Context,
	client pb.UserServiceClient,
	secretName string,
) (*models.UserDB, error) {
	grpcReq := &pb.UserFilterRequest{
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

	return &models.UserDB{
		SecretName: grpcResp.SecretName,
		Username:   grpcResp.Username,
		Password:   grpcResp.Password,
		Meta:       meta,
		UpdatedAt:  updatedAt,
	}, nil
}

func UserAddGRPC(
	ctx context.Context,
	client pb.UserServiceClient,
	req *models.UserAddRequest,
) error {
	grpcReq := &pb.UserAddRequest{
		SecretName: req.SecretName,
		Username:   req.Username,
		Password:   req.Password,
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

func UserAddHTTP(
	ctx context.Context,
	client *resty.Client,
	req *models.UserAddRequest,
) error {
	r, err := client.R().
		SetContext(ctx).
		SetBody(req).
		Post("/user")
	if err != nil {
		return err
	}
	if r.IsError() {
		return errors.New(r.Status())
	}
	return nil
}
