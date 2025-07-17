package client

import (
	"context"
	"errors"
	"fmt"
	"unicode"

	"github.com/jmoiron/sqlx"
	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc"

	"github.com/go-resty/resty/v2"
	"github.com/sbilibin2017/gophkeeper/internal/models"
)

// AuthHTTP sends an HTTP authentication request using models.AuthRequest and models.AuthResponse.
// It posts the request to the "/auth" endpoint.
//
// Returns the AuthResponse containing the authentication token if successful,
// or an error if the request failed or the status code is not 200.
func AuthHTTP(
	ctx context.Context,
	client *resty.Client,
	req *models.AuthRequest,
) (*models.AuthResponse, error) {
	resp := &models.AuthResponse{}

	httpResp, err := client.R().
		SetContext(ctx).
		SetBody(req).
		SetResult(resp).
		Post("/auth")
	if err != nil {
		return nil, fmt.Errorf("failed to send HTTP auth request: %w", err)
	}

	if httpResp.StatusCode() != 200 {
		return nil, fmt.Errorf("auth failed with status %d: %s", httpResp.StatusCode(), httpResp.String())
	}

	return resp, nil
}

// AuthGRPC calls the Auth method of the gRPC AuthService using protobuf request and response.
// It converts the models.AuthRequest to protobuf.AuthRequest, calls the service, and then
// converts the protobuf.AuthResponse back to models.AuthResponse.
//
// Returns the AuthResponse containing the authentication token if successful,
// or an error if the gRPC call failed.
func AuthGRPC(
	ctx context.Context,
	client pb.AuthServiceClient,
	req *models.AuthRequest,
) (*models.AuthResponse, error) {
	pbReq := &pb.AuthRequest{
		Username: req.Username,
		Password: req.Password,
	}

	pbResp, err := client.Auth(ctx, pbReq)
	if err != nil {
		return nil, fmt.Errorf("grpc auth call failed: %w", err)
	}

	resp := &models.AuthResponse{
		Token: pbResp.Token,
	}

	return resp, nil
}

// ValidateRegisterUsername checks if the username meets registration requirements:
// length between 3 and 30 and only letters, digits, or underscore.
//
// Returns an error if validation fails, nil otherwise.
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

// ValidateRegisterPassword checks if the password meets registration requirements:
// at least 8 characters including uppercase, lowercase, and digit.
//
// Returns an error if validation fails, nil otherwise.
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

// ValidateLoginUsername checks that the username is not empty for login.
//
// Returns an error if username is empty, nil otherwise.
func ValidateLoginUsername(username string) error {
	if username == "" {
		return errors.New("username must not be empty")
	}
	return nil
}

// ValidateLoginPassword checks that the password is not empty for login.
//
// Returns an error if password is empty, nil otherwise.
func ValidateLoginPassword(password string) error {
	if password == "" {
		return errors.New("password must not be empty")
	}
	return nil
}

// CreateBinaryRequestTable drops the existing 'secret_binary_request' table if exists,
// then creates a new table with columns: secret_name (primary key), data (binary), and meta.
//
// Returns an error if the operation fails.
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

// CreateTextRequestTable drops the existing 'secret_text_request' table if exists,
// then creates a new table with columns: secret_name (primary key), content (text), and meta.
//
// Returns an error if the operation fails.
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

// CreateUsernamePasswordRequestTable drops the existing 'secret_username_password_request' table if exists,
// then creates a new table with columns: secret_name (primary key), username, password, and meta.
//
// Returns an error if the operation fails.
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

// CreateBankCardRequestTable drops the existing 'secret_bank_card_request' table if exists,
// then creates a new table with columns: secret_name (primary key), number, owner, exp, cvv, and meta.
//
// Returns an error if the operation fails.
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
