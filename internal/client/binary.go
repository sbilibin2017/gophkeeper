package client

import (
	"context"
	"errors"
	"io"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/jmoiron/sqlx"
	"github.com/sbilibin2017/gophkeeper/internal/models"
	"github.com/sbilibin2017/gophkeeper/internal/models/fields"
	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

// BinaryAddClient inserts or updates a binary secret in the local database.
func BinaryAddClient(
	ctx context.Context,
	db *sqlx.DB,
	req *models.BinaryAddRequest,
) error {
	query := `
		INSERT INTO binary_client (secret_name, data, meta, updated_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (secret_name) DO UPDATE SET
			data = EXCLUDED.data,
			meta = EXCLUDED.meta,
			updated_at = EXCLUDED.updated_at
	`

	_, err := db.ExecContext(ctx, query,
		req.SecretName,
		req.Data,
		req.Meta,
		time.Now().UTC(),
	)

	return err
}

// BinaryGetClient retrieves a binary secret from the local database by secret name.
func BinaryGetClient(
	ctx context.Context,
	db *sqlx.DB,
	secretName string,
) (*models.BinaryDB, error) {
	query := `
		SELECT secret_name, data, meta, updated_at
		FROM binary_client
		WHERE secret_name = $1
		LIMIT 1
	`

	var binary models.BinaryDB
	err := db.GetContext(ctx, &binary, query, secretName)
	if err != nil {
		return nil, err
	}
	return &binary, nil
}

// BinaryListClient returns all binary secrets stored in the local database.
func BinaryListClient(
	ctx context.Context,
	db *sqlx.DB,
) ([]models.BinaryDB, error) {
	query := `
		SELECT secret_name, data, meta, updated_at
		FROM binary_client
	`

	var binaries []models.BinaryDB
	err := db.SelectContext(ctx, &binaries, query)
	if err != nil {
		return nil, err
	}

	return binaries, nil
}

// BinaryGetHTTP fetches a binary secret from the remote HTTP API by secret name.
func BinaryGetHTTP(
	ctx context.Context,
	client *resty.Client,
	secretName string,
) (*models.BinaryDB, error) {
	resp := &models.BinaryDB{}
	r, err := client.R().
		SetContext(ctx).
		SetResult(resp).
		Get("/binary/" + secretName)
	if err != nil {
		return nil, err
	}
	if r.IsError() {
		return nil, errors.New(r.Status())
	}
	return resp, nil
}

// BinaryGetGRPC fetches a binary secret from the remote gRPC service by secret name.
func BinaryGetGRPC(
	ctx context.Context,
	client pb.BinaryServiceClient,
	secretName string,
) (*models.BinaryDB, error) {
	grpcReq := &pb.BinaryFilterRequest{
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

	return &models.BinaryDB{
		SecretName: grpcResp.SecretName,
		Data:       grpcResp.Data,
		Meta:       meta,
		UpdatedAt:  updatedAt,
	}, nil
}

// BinaryAddGRPC sends a request to the remote gRPC service to add or update a binary secret.
func BinaryAddGRPC(
	ctx context.Context,
	client pb.BinaryServiceClient,
	req *models.BinaryAddRequest,
) error {
	grpcReq := &pb.BinaryAddRequest{
		SecretName: req.SecretName,
		Data:       req.Data,
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

// BinaryAddHTTP sends a POST request to the remote HTTP API to add or update a binary secret.
func BinaryAddHTTP(
	ctx context.Context,
	client *resty.Client,
	req *models.BinaryAddRequest,
) error {
	r, err := client.R().
		SetContext(ctx).
		SetBody(req).
		Post("/binary")
	if err != nil {
		return err
	}
	if r.IsError() {
		return errors.New(r.Status())
	}
	return nil
}

// BinaryListHTTP fetches a list of all binary secrets from the remote HTTP API.
func BinaryListHTTP(
	ctx context.Context,
	client *resty.Client,
) ([]models.BinaryDB, error) {
	var resp []models.BinaryDB
	r, err := client.R().
		SetContext(ctx).
		SetResult(&resp).
		Get("/binary/")
	if err != nil {
		return nil, err
	}
	if r.IsError() {
		return nil, errors.New(r.Status())
	}
	return resp, nil
}

// BinaryListGRPC fetches a list of all binary secrets from the remote gRPC service using stream.
func BinaryListGRPC(
	ctx context.Context,
	client pb.BinaryServiceClient,
) ([]models.BinaryDB, error) {
	stream, err := client.List(ctx, &emptypb.Empty{})
	if err != nil {
		return nil, err
	}

	var binaries []models.BinaryDB

	for {
		grpcBinary, err := stream.Recv()
		if err == io.EOF {
			break // Stream finished
		}
		if err != nil {
			return nil, err
		}

		var updatedAt time.Time
		if grpcBinary.UpdatedAt != nil {
			updatedAt = grpcBinary.UpdatedAt.AsTime()
		}

		var meta *fields.StringMap
		if grpcBinary.Meta != nil {
			sm := &fields.StringMap{
				Map: make(map[string]string, len(grpcBinary.Meta)),
			}
			for k, v := range grpcBinary.Meta {
				sm.Map[k] = v
			}
			meta = sm
		}

		binary := models.BinaryDB{
			SecretName: grpcBinary.SecretName,
			Data:       grpcBinary.Data,
			Meta:       meta,
			UpdatedAt:  updatedAt,
		}

		binaries = append(binaries, binary)
	}

	return binaries, nil
}
