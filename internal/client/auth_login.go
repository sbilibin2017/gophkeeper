package client

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/jmoiron/sqlx"
	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc"

	"github.com/go-resty/resty/v2"
	"github.com/sbilibin2017/gophkeeper/internal/models"
)

// LoginHTTP sends an HTTP login request to the server.
//
// It takes a context, a Resty HTTP client, and a LoginRequest model,
// and returns a LoginResponse model or an error.
func LoginHTTP(
	ctx context.Context,
	client *resty.Client,
	req *models.LoginRequest,
) (*models.LoginResponse, error) {
	resp := &models.LoginResponse{}

	httpResp, err := client.R().
		SetContext(ctx).
		SetBody(req).
		SetResult(resp).
		Post("/login")
	if err != nil {
		return nil, fmt.Errorf("failed to send HTTP login request: %w", err)
	}

	if httpResp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("login failed with status %d: %s", httpResp.StatusCode(), httpResp.String())
	}

	return resp, nil
}

// LoginGRPC performs user login using a gRPC client.
//
// It takes a context, a gRPC LoginServiceClient, and a LoginRequest model,
// and returns a LoginResponse or an error.
func LoginGRPC(
	ctx context.Context,
	client pb.LoginServiceClient,
	req *models.LoginRequest,
) (*models.LoginResponse, error) {
	pbReq := &pb.LoginRequest{
		Username: req.Username,
		Password: req.Password,
	}

	pbResp, err := client.Login(ctx, pbReq)
	if err != nil {
		return nil, fmt.Errorf("grpc login call failed: %w", err)
	}

	return &models.LoginResponse{Token: pbResp.Token}, nil
}

// ValidateLoginUsername validates that the username is not empty for login.
//
// Returns an error if username is empty.
func ValidateLoginUsername(username string) error {
	if username == "" {
		return errors.New("username must not be empty")
	}
	return nil
}

// ValidateLoginPassword validates that the password is not empty for login.
//
// Returns an error if password is empty.
func ValidateLoginPassword(password string) error {
	if password == "" {
		return errors.New("password must not be empty")
	}
	return nil
}

// CreateBinaryRequestTable creates the "secret_binary_request" table in the database.
func CreateBinaryRequestTable(ctx context.Context, db *sqlx.DB) error {
	_, _ = db.ExecContext(ctx, `DROP TABLE IF EXISTS secret_binary_request;`)
	_, err := db.ExecContext(ctx, `
		CREATE TABLE secret_binary_request (
			secret_name TEXT PRIMARY KEY,
			data BYTEA NOT NULL,
			meta TEXT
		);
	`)
	return err
}

// CreateTextRequestTable creates the "secret_text_request" table in the database.
func CreateTextRequestTable(ctx context.Context, db *sqlx.DB) error {
	_, _ = db.ExecContext(ctx, `DROP TABLE IF EXISTS secret_text_request;`)
	_, err := db.ExecContext(ctx, `
		CREATE TABLE secret_text_request (
			secret_name TEXT PRIMARY KEY,
			content TEXT NOT NULL,
			meta TEXT
		);
	`)
	return err
}

// CreateUsernamePasswordRequestTable creates the "secret_username_password_request" table in the database.
func CreateUsernamePasswordRequestTable(ctx context.Context, db *sqlx.DB) error {
	_, _ = db.ExecContext(ctx, `DROP TABLE IF EXISTS secret_username_password_request;`)
	_, err := db.ExecContext(ctx, `
		CREATE TABLE secret_username_password_request (
			secret_name TEXT PRIMARY KEY,
			username TEXT NOT NULL,
			password TEXT NOT NULL,
			meta TEXT
		);
	`)
	return err
}

// CreateBankCardRequestTable creates the "secret_bank_card_request" table in the database.
func CreateBankCardRequestTable(ctx context.Context, db *sqlx.DB) error {
	_, _ = db.ExecContext(ctx, `DROP TABLE IF EXISTS secret_bank_card_request;`)
	_, err := db.ExecContext(ctx, `
		CREATE TABLE secret_bank_card_request (
			secret_name TEXT PRIMARY KEY,
			number TEXT NOT NULL,
			owner TEXT,
			exp TEXT NOT NULL,
			cvv TEXT NOT NULL,
			meta TEXT
		);
	`)
	return err
}
