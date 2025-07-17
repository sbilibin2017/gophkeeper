package client

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"unicode"

	"github.com/go-resty/resty/v2"
	"github.com/sbilibin2017/gophkeeper/internal/models"
	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc"
)

// RegisterHTTP sends an HTTP registration request to the server.
func RegisterHTTP(
	ctx context.Context,
	client *resty.Client,
	req *models.RegisterRequest,
) error {
	httpResp, err := client.R().
		SetContext(ctx).
		SetBody(req).
		Post("/register")
	if err != nil {
		return fmt.Errorf("failed to send HTTP register request: %w", err)
	}

	if httpResp.StatusCode() != http.StatusOK {
		return fmt.Errorf("registration failed with status %d: %s", httpResp.StatusCode(), httpResp.String())
	}

	return nil
}

// RegisterGRPC performs user registration using a gRPC client.
func RegisterGRPC(
	ctx context.Context,
	client pb.RegisterServiceClient,
	req *models.RegisterRequest,
) error {
	pbReq := &pb.RegisterRequest{
		Username: req.Username,
		Password: req.Password,
	}

	_, err := client.Register(ctx, pbReq)
	if err != nil {
		return fmt.Errorf("grpc register call failed: %w", err)
	}

	return nil
}

// ValidateRegisterUsername ensures the username is valid.
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

// ValidateRegisterPassword ensures the password meets strength requirements.
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
