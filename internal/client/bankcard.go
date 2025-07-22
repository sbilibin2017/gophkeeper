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

// BankCardAddClient inserts or updates a bank card in the DB.
func BankCardAddClient(
	ctx context.Context,
	db *sqlx.DB,
	req *models.BankCardAddRequest,
) error {
	query := `
		INSERT INTO bankcard_client (secret_name, number, owner, exp, cvv, meta, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (secret_name) DO UPDATE SET
			number = EXCLUDED.number,
			owner = EXCLUDED.owner,
			exp = EXCLUDED.exp,
			cvv = EXCLUDED.cvv,
			meta = EXCLUDED.meta,
			updated_at = EXCLUDED.updated_at
	`

	_, err := db.ExecContext(ctx, query,
		req.SecretName,
		req.Number,
		req.Owner,
		req.Exp,
		req.CVV,
		req.Meta, // *fields.StringMap implements driver.Valuer
		time.Now().UTC(),
	)
	return err
}

// BankCardGetClient retrieves a bank card from the DB by secret name.
func BankCardGetClient(
	ctx context.Context,
	db *sqlx.DB,
	secretName string,
) (*models.BankCardDB, error) {
	query := `
		SELECT secret_name, number, owner, exp, cvv, meta, updated_at
		FROM bankcard_client
		WHERE secret_name = $1
		LIMIT 1
	`

	var card models.BankCardDB
	err := db.GetContext(ctx, &card, query, secretName)
	if err != nil {
		return nil, err
	}
	return &card, nil
}

// BankCardListClient retrieves all bank cards from the DB.
func BankCardListClient(
	ctx context.Context,
	db *sqlx.DB,
) ([]models.BankCardDB, error) {
	query := `
		SELECT secret_name, number, owner, exp, cvv, meta, updated_at
		FROM bankcard_client
	`

	var cards []models.BankCardDB
	err := db.SelectContext(ctx, &cards, query)
	if err != nil {
		return nil, err
	}

	return cards, nil
}

// BankCardGetHTTP gets a bank card via HTTP.
func BankCardGetHTTP(
	ctx context.Context,
	client *resty.Client,
	secretName string,
) (*models.BankCardDB, error) {
	resp := &models.BankCardDB{}
	r, err := client.R().
		SetContext(ctx).
		SetResult(resp).
		Get("/bankcard/" + secretName)
	if err != nil {
		return nil, err
	}
	if r.IsError() {
		return nil, errors.New(r.Status())
	}
	return resp, nil
}

func BankCardGetGRPC(
	ctx context.Context,
	client pb.BankCardServiceClient,
	secretName string,
) (*models.BankCardDB, error) {
	grpcReq := &pb.BankCardFilterRequest{
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

	return &models.BankCardDB{
		SecretName:  grpcResp.SecretName,
		SecretOwner: grpcResp.SecretOwner,
		Number:      grpcResp.Number,
		Owner:       grpcResp.Owner,
		Exp:         grpcResp.Exp,
		CVV:         grpcResp.Cvv,
		Meta:        meta,
		UpdatedAt:   updatedAt,
	}, nil
}

// BankCardAddHTTP adds a bank card via HTTP.
func BankCardAddHTTP(
	ctx context.Context,
	client *resty.Client,
	req *models.BankCardAddRequest,
) error {
	r, err := client.R().
		SetContext(ctx).
		SetBody(req).
		Post("/bankcard")
	if err != nil {
		return err
	}
	if r.IsError() {
		return errors.New(r.Status())
	}
	return nil
}

// BankCardAddGRPC adds a bank card via gRPC.
func BankCardAddGRPC(
	ctx context.Context,
	client pb.BankCardServiceClient,
	req *models.BankCardAddRequest,
) error {
	grpcReq := &pb.BankCardAddRequest{
		SecretName: req.SecretName,
		Number:     req.Number,
		Owner:      req.Owner,
		Exp:        req.Exp,
		Cvv:        req.CVV,
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
