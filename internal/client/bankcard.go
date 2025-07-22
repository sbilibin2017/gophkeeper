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

// BankCardAddClient inserts a bank card into the local database,
// or updates it if the secret name already exists.
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
		req.Meta,
		time.Now().UTC(),
	)
	return err
}

// BankCardGetClient retrieves a bank card from the local database
// by its secret name.
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

// BankCardListClient returns a list of all bank cards
// stored in the local database.
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

// BankCardListHTTP fetches a list of all bank cards from the remote HTTP API.
func BankCardListHTTP(
	ctx context.Context,
	client *resty.Client,
) ([]models.BankCardDB, error) {
	var resp []models.BankCardDB
	r, err := client.R().
		SetContext(ctx).
		SetResult(&resp).
		Get("/bankcard/")
	if err != nil {
		return nil, err
	}
	if r.IsError() {
		return nil, errors.New(r.Status())
	}
	return resp, nil
}

// BankCardListGRPC fetches a list of all bank cards from the remote gRPC service using stream.
func BankCardListGRPC(
	ctx context.Context,
	client pb.BankCardServiceClient,
) ([]models.BankCardDB, error) {
	stream, err := client.List(ctx, &emptypb.Empty{})
	if err != nil {
		return nil, err
	}

	var cards []models.BankCardDB

	for {
		grpcCard, err := stream.Recv()
		if err == io.EOF {
			break // Stream finished
		}
		if err != nil {
			return nil, err
		}

		var updatedAt time.Time
		if grpcCard.UpdatedAt != nil {
			updatedAt = grpcCard.UpdatedAt.AsTime()
		}

		var meta *fields.StringMap
		if grpcCard.Meta != nil {
			sm := &fields.StringMap{
				Map: make(map[string]string, len(grpcCard.Meta)),
			}
			for k, v := range grpcCard.Meta {
				sm.Map[k] = v
			}
			meta = sm
		}

		card := models.BankCardDB{
			SecretName:  grpcCard.SecretName,
			SecretOwner: grpcCard.SecretOwner,
			Number:      grpcCard.Number,
			Owner:       grpcCard.Owner,
			Exp:         grpcCard.Exp,
			CVV:         grpcCard.Cvv,
			Meta:        meta,
			UpdatedAt:   updatedAt,
		}

		cards = append(cards, card)
	}

	return cards, nil
}

// BankCardGetHTTP fetches a bank card from the remote HTTP API
// by secret name using a GET request.
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

// BankCardGetGRPC fetches a bank card from the remote gRPC service
// by secret name.
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

// BankCardAddHTTP sends a POST request to the remote HTTP API
// to add or update a bank card.
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

// BankCardAddGRPC sends a gRPC request to add or update a bank card
// on the remote service.
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
