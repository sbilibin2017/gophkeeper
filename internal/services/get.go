package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/sbilibin2017/gophkeeper/internal/models"
	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc"
)

// GetHTTPLoginPassword fetches a LoginPassword secret via HTTP using Resty client.
// It requires a context, Resty client, auth token, and secret ID.
// Returns the LoginPassword model or an error.
func GetHTTPLoginPassword(
	ctx context.Context,
	client *resty.Client,
	token string,
	secretID string,
) (*models.LoginPassword, error) {
	const op = "services.GetHTTPLoginPassword"

	url := "/login-password/" + secretID

	resp, err := client.R().
		SetContext(ctx).
		SetHeader("Authorization", "Bearer "+token).
		SetHeader("Accept", "application/json").
		Get(url)
	if err != nil {
		return nil, fmt.Errorf("%s: http request failed: %w", op, err)
	}
	if resp.IsError() {
		return nil, fmt.Errorf("%s: server returned error: status %d, body: %s", op, resp.StatusCode(), resp.String())
	}

	var result models.LoginPassword
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, fmt.Errorf("%s: unmarshal loginpassword response failed: %w", op, err)
	}

	return &result, nil
}

// GetHTTPText fetches a Text secret via HTTP using Resty client.
// It requires a context, Resty client, auth token, and secret ID.
// Returns the Text model or an error.
func GetHTTPText(
	ctx context.Context,
	client *resty.Client,
	token,
	secretID string,
) (*models.Text, error) {
	const op = "services.GetHTTPText"

	url := "/text/" + secretID

	resp, err := client.R().
		SetContext(ctx).
		SetHeader("Authorization", "Bearer "+token).
		SetHeader("Accept", "application/json").
		Get(url)
	if err != nil {
		return nil, fmt.Errorf("%s: http request failed: %w", op, err)
	}
	if resp.IsError() {
		return nil, fmt.Errorf("%s: server returned error: status %d, body: %s", op, resp.StatusCode(), resp.String())
	}

	var result models.Text
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, fmt.Errorf("%s: unmarshal text response failed: %w", op, err)
	}

	return &result, nil
}

// GetHTTPBinary fetches a Binary secret via HTTP using Resty client.
// It requires a context, Resty client, auth token, and secret ID.
// Returns the Binary model or an error.
func GetHTTPBinary(
	ctx context.Context,
	client *resty.Client,
	token string,
	secretID string,
) (*models.Binary, error) {
	const op = "services.GetHTTPBinary"

	resp, err := client.R().
		SetContext(ctx).
		SetHeader("Authorization", "Bearer "+token).
		SetHeader("Accept", "application/json").
		Get("/binary/" + secretID)
	if err != nil {
		return nil, fmt.Errorf("%s: http request failed: %w", op, err)
	}
	if resp.IsError() {
		return nil, fmt.Errorf("%s: server returned error: status %d, body: %s", op, resp.StatusCode(), resp.String())
	}

	var result models.Binary
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, fmt.Errorf("%s: unmarshal binary response failed: %w", op, err)
	}

	return &result, nil
}

// GetHTTPCard fetches a Card secret via HTTP using Resty client.
// It requires a context, Resty client, auth token, and secret ID.
// Returns the Card model or an error.
func GetHTTPCard(
	ctx context.Context,
	client *resty.Client,
	token string,
	secretID string,
) (*models.Card, error) {
	const op = "services.GetHTTPCard"

	url := "/card/" + secretID

	resp, err := client.R().
		SetContext(ctx).
		SetHeader("Authorization", "Bearer "+token).
		SetHeader("Accept", "application/json").
		Get(url)
	if err != nil {
		return nil, fmt.Errorf("%s: http request failed: %w", op, err)
	}
	if resp.IsError() {
		return nil, fmt.Errorf("%s: server returned error: status %d, body: %s", op, resp.StatusCode(), resp.String())
	}

	var result models.Card
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, fmt.Errorf("%s: unmarshal card response failed: %w", op, err)
	}

	return &result, nil
}

// GetGRPCLoginPassword fetches a LoginPassword secret via gRPC call.
// It requires context, gRPC client, auth token, and secret ID.
// Returns the LoginPassword model or an error.
func GetGRPCLoginPassword(
	ctx context.Context,
	client pb.GetLoginPasswordServiceClient,
	token string,
	secretID string,
) (*models.LoginPassword, error) {
	const op = "services.GetLoginPassword"

	req := &pb.GetRequest{
		Token:    token,
		SecretId: secretID,
	}

	resp, err := client.Get(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("%s: gRPC call failed: %w", op, err)
	}
	if resp.Error != "" {
		return nil, fmt.Errorf("%s: server returned error: %s", op, resp.Error)
	}

	lp := resp.LoginPassword
	return &models.LoginPassword{
		SecretID:  lp.SecretId,
		Login:     lp.Login,
		Password:  lp.Password,
		Meta:      lp.Meta,
		UpdatedAt: time.Unix(lp.UpdatedAt, 0),
	}, nil
}

// GetGRPCText fetches a Text secret via gRPC call.
// It requires context, gRPC client, auth token, and secret ID.
// Returns the Text model or an error.
func GetGRPCText(
	ctx context.Context,
	client pb.GetTextServiceClient,
	token,
	secretID string,
) (*models.Text, error) {
	const op = "services.GetText"

	req := &pb.GetRequest{
		Token:    token,
		SecretId: secretID,
	}

	resp, err := client.Get(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("%s: gRPC call failed: %w", op, err)
	}
	if resp.Error != "" {
		return nil, fmt.Errorf("%s: server returned error: %s", op, resp.Error)
	}

	t := resp.Text
	return &models.Text{
		SecretID:  t.SecretId,
		Content:   t.Content,
		Meta:      t.Meta,
		UpdatedAt: time.Unix(t.UpdatedAt, 0),
	}, nil
}

// GetGRPCBinary fetches a Binary secret via gRPC call.
// It requires context, gRPC client, auth token, and secret ID.
// Returns the Binary model or an error.
func GetGRPCBinary(
	ctx context.Context,
	client pb.GetBinaryServiceClient,
	token string,
	secretID string,
) (*models.Binary, error) {
	const op = "services.GetBinary"

	req := &pb.GetRequest{
		Token:    token,
		SecretId: secretID,
	}

	resp, err := client.Get(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("%s: gRPC call failed: %w", op, err)
	}
	if resp.Error != "" {
		return nil, fmt.Errorf("%s: server returned error: %s", op, resp.Error)
	}

	b := resp.Binary
	return &models.Binary{
		SecretID:  b.SecretId,
		Data:      b.Data,
		Meta:      b.Meta,
		UpdatedAt: time.Unix(b.UpdatedAt, 0),
	}, nil
}

// GetGRPCCard fetches a Card secret via gRPC call.
// It requires context, gRPC client, auth token, and secret ID.
// Returns the Card model or an error.
func GetGRPCCard(
	ctx context.Context,
	client pb.GetCardServiceClient,
	token string,
	secretID string,
) (*models.Card, error) {
	const op = "services.GetCard"

	req := &pb.GetRequest{
		Token:    token,
		SecretId: secretID,
	}

	resp, err := client.Get(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("%s: gRPC call failed: %w", op, err)
	}
	if resp.Error != "" {
		return nil, fmt.Errorf("%s: server returned error: %s", op, resp.Error)
	}

	c := resp.Card
	return &models.Card{
		SecretID:  c.SecretId,
		Number:    c.Number,
		Holder:    c.Holder,
		ExpMonth:  int(c.ExpMonth),
		ExpYear:   int(c.ExpYear),
		CVV:       c.Cvv,
		Meta:      c.Meta,
		UpdatedAt: time.Unix(c.UpdatedAt, 0),
	}, nil
}
