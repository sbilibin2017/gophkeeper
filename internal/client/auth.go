package client

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"unicode"

	"github.com/jmoiron/sqlx"
	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc"

	"github.com/go-resty/resty/v2"
	"github.com/sbilibin2017/gophkeeper/internal/models"
)

// RegisterHTTP sends an HTTP registration request to the server.
//
// It takes a context, a Resty HTTP client, and a RegisterRequest model,
// and returns a RegisterResponse model or an error.
func RegisterHTTP(
	ctx context.Context,
	client *resty.Client,
	req *models.RegisterRequest,
) (*models.RegisterResponse, error) {
	resp := &models.RegisterResponse{}

	httpResp, err := client.R().
		SetContext(ctx).
		SetBody(req).
		SetResult(resp).
		Post("/register")
	if err != nil {
		return nil, fmt.Errorf("failed to send HTTP register request: %w", err)
	}

	if httpResp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("registration failed with status %d: %s", httpResp.StatusCode(), httpResp.String())
	}

	return resp, nil
}

// RegisterGRPC performs user registration using a gRPC client.
//
// It takes a context, a gRPC RegisterServiceClient, and a RegisterRequest model,
// and returns a RegisterResponse or an error.
func RegisterGRPC(
	ctx context.Context,
	client pb.RegisterServiceClient,
	req *models.RegisterRequest,
) (*models.RegisterResponse, error) {
	pbReq := &pb.RegisterRequest{
		Username: req.Username,
		Password: req.Password,
	}

	pbResp, err := client.Register(ctx, pbReq)
	if err != nil {
		return nil, fmt.Errorf("grpc register call failed: %w", err)
	}

	return &models.RegisterResponse{Token: pbResp.Token}, nil
}

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

// LogoutHTTP sends an HTTP logout request.
//
// It takes a context, a Resty HTTP client, and a LogoutRequest model.
// Returns an error if the logout failed.
func LogoutHTTP(ctx context.Context, client *resty.Client, req *models.LogoutRequest) error {
	resp, err := client.R().
		SetContext(ctx).
		SetHeader("Authorization", req.Token).
		Post("/logout")
	if err != nil {
		return err
	}

	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("logout failed: %s", resp.Status())
	}

	return nil
}

// LogoutGRPC performs user logout using a gRPC client.
//
// It takes a context, a gRPC LogoutServiceClient, and a LogoutRequest model.
// Returns an error if the logout failed.
func LogoutGRPC(
	ctx context.Context,
	client pb.LogoutServiceClient,
	req *models.LogoutRequest,
) error {
	_, err := client.Logout(ctx, &pb.LogoutRequest{
		Token: req.Token,
	})
	if err != nil {
		return fmt.Errorf("grpc logout call failed: %w", err)
	}
	return nil
}

// ValidateRegisterUsername validates if the username meets registration criteria.
//
// Username must be between 3 and 30 characters and contain only letters, digits, or underscore.
// Returns an error if validation fails.
func ValidateRegisterUsername(username string) error {
	if len(username) < 3 || len(username) > 30 {
		return errors.New("username must be between 3 and 30 characters")
	}
	for _, ch := range username {
		if !(unicode.IsLetter(ch) || unicode.IsDigit(ch) || ch == '_') {
			return errors.New("username can only contain letters, digits, and underscore")
		}
	}
	return nil
}

// ValidateRegisterPassword validates if the password meets registration criteria.
//
// Password must be at least 8 characters long and contain at least one uppercase letter,
// one lowercase letter, and one digit. Returns an error if validation fails.
func ValidateRegisterPassword(password string) error {
	if len(password) < 8 {
		return errors.New("password must be at least 8 characters long")
	}
	var hasUpper, hasLower, hasDigit bool
	for _, ch := range password {
		switch {
		case unicode.IsUpper(ch):
			hasUpper = true
		case unicode.IsLower(ch):
			hasLower = true
		case unicode.IsDigit(ch):
			hasDigit = true
		}
	}
	if !hasUpper {
		return errors.New("password must contain at least one uppercase letter")
	}
	if !hasLower {
		return errors.New("password must contain at least one lowercase letter")
	}
	if !hasDigit {
		return errors.New("password must contain at least one digit")
	}
	return nil
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
//
// Drops the table if it exists and then creates it with columns secret_name (primary key),
// data (binary), and meta (text).
func CreateBinaryRequestTable(db *sqlx.DB) error {
	_, _ = db.Exec(`DROP TABLE IF EXISTS secret_binary_request;`)
	_, err := db.Exec(`
		CREATE TABLE secret_binary_request (
			secret_name TEXT PRIMARY KEY,
			data BYTEA NOT NULL,
			meta TEXT
		);
	`)
	return err
}

// CreateTextRequestTable creates the "secret_text_request" table in the database.
//
// Drops the table if it exists and then creates it with columns secret_name (primary key),
// content (text), and meta (text).
func CreateTextRequestTable(db *sqlx.DB) error {
	_, _ = db.Exec(`DROP TABLE IF EXISTS secret_text_request;`)
	_, err := db.Exec(`
		CREATE TABLE secret_text_request (
			secret_name TEXT PRIMARY KEY,
			content TEXT NOT NULL,
			meta TEXT
		);
	`)
	return err
}

// CreateUsernamePasswordRequestTable creates the "secret_username_password_request" table in the database.
//
// Drops the table if it exists and then creates it with columns secret_name (primary key),
// username, password, and meta (all text).
func CreateUsernamePasswordRequestTable(db *sqlx.DB) error {
	_, _ = db.Exec(`DROP TABLE IF EXISTS secret_username_password_request;`)
	_, err := db.Exec(`
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
//
// Drops the table if it exists and then creates it with columns secret_name (primary key),
// number, owner, exp, cvv, and meta (all text except owner optional).
func CreateBankCardRequestTable(db *sqlx.DB) error {
	_, _ = db.Exec(`DROP TABLE IF EXISTS secret_bank_card_request;`)
	_, err := db.Exec(`
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
