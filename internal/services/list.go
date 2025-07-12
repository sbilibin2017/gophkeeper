package services

import (
	"context"
	"fmt"

	"github.com/go-resty/resty/v2"
	"github.com/sbilibin2017/gophkeeper/internal/models"
	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc"
)

func ListUsernamePasswordHTTP(
	ctx context.Context,
	client *resty.Client,
) ([]*models.UsernamePassword, error) {
	var result []*models.UsernamePassword
	resp, err := client.R().
		SetContext(ctx).
		SetResult(&result).
		Get("/list/username-passwords")

	if err != nil {
		return nil, fmt.Errorf("ошибка запроса: %w", err)
	}

	if resp.IsError() {
		return nil, fmt.Errorf("ошибка сервера: %s", resp.Status())
	}

	return result, nil
}

// ListUsernamePasswordGRPC получает список UsernamePassword через gRPC клиента.
func ListUsernamePasswordGRPC(
	ctx context.Context,
	client pb.ListServiceClient,
) ([]*pb.UsernamePasswordItem, error) {
	req := &pb.ListUsernamePasswordRequest{}

	resp, err := client.ListUsernamePassword(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to perform gRPC request: %w", err)
	}

	if resp.Error != "" {
		return nil, fmt.Errorf("server error: %s", resp.Error)
	}

	return resp.Items, nil
}

func ListTextHTTP(
	ctx context.Context,
	client *resty.Client,
) ([]*models.Text, error) {
	var result []*models.Text
	resp, err := client.R().
		SetContext(ctx).
		SetResult(&result).
		Get("/list/texts")

	if err != nil {
		return nil, fmt.Errorf("ошибка запроса: %w", err)
	}

	if resp.IsError() {
		return nil, fmt.Errorf("ошибка сервера: %s", resp.Status())
	}

	return result, nil
}

// ListTextGRPC получает список текстов через gRPC клиента.
func ListTextGRPC(
	ctx context.Context,
	client pb.ListServiceClient,
) ([]*pb.TextItem, error) {
	req := &pb.ListTextRequest{}

	resp, err := client.ListText(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to perform gRPC request: %w", err)
	}

	if resp.Error != "" {
		return nil, fmt.Errorf("server error: %s", resp.Error)
	}

	return resp.Items, nil
}

func ListBinaryHTTP(
	ctx context.Context,
	client *resty.Client,
) ([]*models.Binary, error) {
	var result []*models.Binary
	resp, err := client.R().
		SetContext(ctx).
		SetResult(&result).
		Get("/list/binaries")

	if err != nil {
		return nil, fmt.Errorf("ошибка запроса: %w", err)
	}

	if resp.IsError() {
		return nil, fmt.Errorf("ошибка сервера: %s", resp.Status())
	}

	return result, nil
}

// ListBinaryGRPC получает список бинарных данных через gRPC клиента.
func ListBinaryGRPC(
	ctx context.Context,
	client pb.ListServiceClient,
) ([]*pb.BinaryItem, error) {
	req := &pb.ListBinaryRequest{}

	resp, err := client.ListBinary(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to perform gRPC request: %w", err)
	}

	if resp.Error != "" {
		return nil, fmt.Errorf("server error: %s", resp.Error)
	}

	return resp.Items, nil
}

func ListBankCardHTTP(
	ctx context.Context,
	client *resty.Client,
) ([]*models.BankCard, error) {
	var result []*models.BankCard
	resp, err := client.R().
		SetContext(ctx).
		SetResult(&result).
		Get("/list/bank-cards")

	if err != nil {
		return nil, fmt.Errorf("ошибка запроса: %w", err)
	}

	if resp.IsError() {
		return nil, fmt.Errorf("ошибка сервера: %s", resp.Status())
	}

	return result, nil
}

// ListBankCardGRPC получает список банковских карт через gRPC клиента.
func ListBankCardGRPC(
	ctx context.Context,
	client pb.ListServiceClient,
) ([]*pb.BankCardItem, error) {
	req := &pb.ListBankCardRequest{}

	resp, err := client.ListBankCard(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to perform gRPC request: %w", err)
	}

	if resp.Error != "" {
		return nil, fmt.Errorf("server error: %s", resp.Error)
	}

	return resp.Items, nil
}
