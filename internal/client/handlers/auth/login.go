package handlers

import (
	"context"
	"fmt"

	"github.com/sbilibin2017/gophkeeper/internal/client/config"
	"github.com/sbilibin2017/gophkeeper/internal/client/facades/auth"
	"github.com/sbilibin2017/gophkeeper/internal/client/models"
	"github.com/sbilibin2017/gophkeeper/internal/client/repositories/bankcard"
	"github.com/sbilibin2017/gophkeeper/internal/client/repositories/binary"
	"github.com/sbilibin2017/gophkeeper/internal/client/repositories/text"
	"github.com/sbilibin2017/gophkeeper/internal/client/repositories/user"
	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc/auth"
)

func LoginHTTP(
	ctx context.Context,
	username, password, authURL, tlsCertFile, tlsKeyFile string,
) (*models.AuthResponse, error) {
	config, err := config.NewConfig(
		config.WithHTTPClient(authURL, tlsCertFile, tlsKeyFile, ""),
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

	facade := auth.NewLoginHTTPFacade(config.HTTPClient)

	resp, err := facade.Login(ctx, authReq)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func LoginGRPC(
	ctx context.Context,
	username, password, authURL, tlsCertFile, tlsKeyFile string,
) (*models.AuthResponse, error) {
	config, err := config.NewConfig(
		config.WithGRPCClient(authURL, tlsCertFile, tlsKeyFile, ""),
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
	facade := auth.NewLoginGRPCFacade(grpcClient)

	resp, err := facade.Login(ctx, authReq)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
