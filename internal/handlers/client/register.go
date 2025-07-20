package client

import (
	"context"
	"fmt"

	"github.com/sbilibin2017/gophkeeper/internal/configs"
	"github.com/sbilibin2017/gophkeeper/internal/facades/auth"
	"github.com/sbilibin2017/gophkeeper/internal/models"
	"github.com/sbilibin2017/gophkeeper/internal/repositories/bankcard"
	"github.com/sbilibin2017/gophkeeper/internal/repositories/binary"
	"github.com/sbilibin2017/gophkeeper/internal/repositories/text"
	"github.com/sbilibin2017/gophkeeper/internal/repositories/user"

	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc/auth"
)

func RegisterHTTP(
	ctx context.Context,
	username string,
	password string,
	authURL string,
	tlsCertFile string,
	tlsKeyFile string,
) (*models.AuthResponse, error) {
	config, err := configs.NewClientConfig(
		configs.WithHTTPClient(authURL, tlsCertFile, tlsKeyFile, ""),
	)
	if err != nil {
		return nil, err
	}
	if config.HTTPClient == nil {
		return nil, fmt.Errorf("HTTP client is not configured for URL: %s", authURL)
	}

	bankcard.CreateClientTable(ctx, config.DB)
	text.CreateClientTable(ctx, config.DB)
	binary.CreateClientTable(ctx, config.DB)
	user.CreateClientTable(ctx, config.DB)

	authReq := &models.AuthRequest{
		Username: username,
		Password: password,
	}

	facade := auth.NewRegisterHTTPFacade(config.HTTPClient)

	resp, err := facade.Register(ctx, authReq)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func RegisterGRPC(
	ctx context.Context,
	username string,
	password string,
	authURL string,
	tlsCertFile string,
	tlsKeyFile string,
) (*models.AuthResponse, error) {
	config, err := configs.NewClientConfig(
		configs.WithGRPCClient(authURL, tlsCertFile, tlsKeyFile, ""),
	)
	if err != nil {
		return nil, err
	}
	if config.GRPCClient == nil {
		return nil, fmt.Errorf("gRPC client is not configured for URL: %s", authURL)
	}

	bankcard.CreateClientTable(ctx, config.DB)
	text.CreateClientTable(ctx, config.DB)
	binary.CreateClientTable(ctx, config.DB)
	user.CreateClientTable(ctx, config.DB)

	authReq := &models.AuthRequest{
		Username: username,
		Password: password,
	}

	grpcClient := pb.NewAuthServiceClient(config.GRPCClient)
	facade := auth.NewRegisterGRPCFacade(grpcClient)

	resp, err := facade.Register(ctx, authReq)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
